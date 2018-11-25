package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/file"
)

const (
	postType = "Post"
)

// MarkdownDocument is a text/markdown document
type MarkdownDocument struct {
	Source string `json:"source"`
}

// NewMarkdownDocument makes a Markdown
func NewMarkdownDocument(source string) MarkdownDocument {
	markdownDocument := MarkdownDocument{
		Source: source,
	}
	return markdownDocument
}

// Post has a markdown document
type Post struct {
	CreatedAt     time.Time      `json:"createdAt"`
	Key           *datastore.Key `datastore:",omitempty" json:"id"`
	ParentPostKey *datastore.Key `json:"parentPostID"`
	//AuthorID string            `json:"authorID"`
	Content           MarkdownDocument `json:"content"`
	ContentStorageKey string           `json:"-"`
	Replies           *[]Post          `datastore:"-" json:"replies,omitempty"`
	CommandType       string           `json:"commandType"`
}

// CreatePostInput is used to create new posts
type CreatePostInput struct {
	ChannelSlug          string
	ParentPostKeyEncoded *string
	MarkdownSource       string
	CommandType          string
}

func objectForPostContentStorage(ctx context.Context, contentStorageKey string) (*storage.ObjectHandle, error) {
	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		return nil, fmt.Errorf("No default bucket found")
	}
	log.Printf("BUCKET NAME %s\n", bucketName)

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot make client to post content storage")
	}

	bucket := storageClient.Bucket(bucketName)
	object := bucket.Object(contentStorageKey)

	return object, nil
}

// CreatePost creates a new post
func (repo ChannelsRepo) CreatePost(input CreatePostInput) (*Post, error) {
	if input.MarkdownSource == "" {
		return nil, fmt.Errorf("Post content cannot be empty")
	}

	var err error

	channelContentKey := repo.channelContentKeyFor(input.ChannelSlug)
	if channelContentKey == nil {
		return nil, fmt.Errorf("No channel with slug: %s", input.ChannelSlug)
	}

	postKey := datastore.NewIncompleteKey(repo.ctx, postType, channelContentKey)

	var parentPostKey *datastore.Key
	if input.ParentPostKeyEncoded != nil {
		parentPostKey, err = datastore.DecodeKey(*input.ParentPostKeyEncoded)
		if err != nil {
			return nil, fmt.Errorf("Invalid parent post id: %s", *input.ParentPostKeyEncoded)
		}
	}

	markdownDocument := NewMarkdownDocument(input.MarkdownSource)
	post := Post{
		ParentPostKey: parentPostKey,
		Content:       markdownDocument,
		CreatedAt:     time.Now().UTC(),
		CommandType:   input.CommandType,
	}

	if len(markdownDocument.Source) >= 1500 {
		i, _, err := datastore.AllocateIDs(repo.ctx, "PostContentStorageKey", channelContentKey, 1)
		if err != nil {
			return nil, fmt.Errorf("Cannot allocate ID for post content")
		}

		contentStorageKey := channelContentKey.Encode() + string(i)
		object, err := objectForPostContentStorage(repo.ctx, contentStorageKey)
		if err != nil {
			return nil, err
		}

		writer := object.NewWriter(repo.ctx)
		writer.ContentType = "text/markdown"
		_, writeErr := io.WriteString(writer, post.Content.Source)
		if writeErr != nil {
			return nil, writeErr
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}

		post.Content.Source = ""
		post.ContentStorageKey = contentStorageKey
	}

	postKey, err = datastore.Put(repo.ctx, postKey, &post)
	if err != nil {
		return nil, err
	}

	post.Key = postKey

	return &post, nil
}

func readPostContentFromStorageIfNeeded(ctx context.Context, post *Post) {
	if post.ContentStorageKey == "" {
		return
	}

	object, err := objectForPostContentStorage(ctx, post.ContentStorageKey)
	if err != nil {
		return
	}

	r, err := object.NewReader(ctx)
	if err != nil {
		return
	}

	bytes, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return
	}

	post.Content.Source = string(bytes)
}

// GetPostWithIDInChannel lists all post in a channel of a certain slug
func (repo ChannelsRepo) GetPostWithIDInChannel(channelSlug string, postID string) (*Post, error) {
	postKey, err := datastore.DecodeKey(postID)

	var post Post
	err = datastore.Get(repo.ctx, postKey, &post)
	if err == datastore.ErrNoSuchEntity {
		return nil, errors.New("No post with id: " + postID)
	}
	if err != nil {
		return nil, errors.New("Error reading post with id: " + postID + ": " + err.Error())
	}

	post.Key = postKey

	readPostContentFromStorageIfNeeded(repo.ctx, &post)

	return &post, nil
}

// NewPostsConnection makes a new connection with the posts in a specific channel
func (repo ChannelsRepo) NewPostsConnection(options PostsConnectionOptions) *PostsConnection {
	connection := PostsConnection{repo: repo, options: options}
	return &connection
}

// ListPostsInChannel lists all post in a channel of a certain slug
func (repo ChannelsRepo) ListPostsInChannel(channelSlug string) ([]Post, error) {
	connection := repo.NewPostsConnection(PostsConnectionOptions{
		channelSlug:    channelSlug,
		includeReplies: true,
		maxCount:       100,
	})

	posts, err := connection.All()
	return posts, err
}
