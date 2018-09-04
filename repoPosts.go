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
	channelSlugType = "ChannelSlug"
	postType        = "Post"
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

// ChannelContent holds main data of a channel
type ChannelContent struct {
	Key         *datastore.Key `datastore:"-" json:"id"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
}

// ChannelSlug allows a channel to be found by slug
type ChannelSlug struct {
	ContentKey *datastore.Key
}

// Post has a markdown document
type Post struct {
	CreatedAt     time.Time      `json:"createdAt"`
	Key           *datastore.Key `datastore:",omitempty" json:"id"`
	ParentPostKey *datastore.Key `json:"parentPostID"`
	//AuthorID string            `json:"authorID"`
	Content           MarkdownDocument `json:"content"`
	ContentStorageKey string
	Replies           *[]Post `datastore:"-" json:"replies"`
}

// ChannelsRepo lets you query the channels repository
type ChannelsRepo struct {
	ctx     context.Context
	orgRepo OrgRepo
}

// NewChannelsRepo makes a new channels repository with the given org name
func NewChannelsRepo(ctx context.Context, orgRepo OrgRepo) ChannelsRepo {
	return ChannelsRepo{
		ctx:     ctx,
		orgRepo: orgRepo,
	}
}

func (repo ChannelsRepo) channelSlugKeyFor(slug string) *datastore.Key {
	return datastore.NewKey(repo.ctx, channelSlugType, slug, 0, repo.orgRepo.RootKey())
}

func (repo ChannelsRepo) channelContentKeyFor(slug string) *datastore.Key {
	channelSlugKey := repo.channelSlugKeyFor(slug)
	var channelSlug = ChannelSlug{}
	err := datastore.Get(repo.ctx, channelSlugKey, &channelSlug)
	if err != nil {
		return nil
	}

	return channelSlug.ContentKey
}

// CreateChannel creates a new channel
func (repo ChannelsRepo) CreateChannel(slug string) (ChannelContent, error) {
	channelSlugKey := repo.channelSlugKeyFor(slug)

	channelContentKey := datastore.NewIncompleteKey(repo.ctx, "ChannelContent", repo.orgRepo.RootKey())
	channelContent := ChannelContent{
		Slug:        slug,
		Description: "",
	}
	channelContentKey, err := datastore.Put(repo.ctx, channelContentKey, &channelContent)

	channelSlug := ChannelSlug{
		ContentKey: channelContentKey,
	}
	_, err = datastore.Put(repo.ctx, channelSlugKey, &channelSlug)

	channelContent.Key = channelContentKey
	return channelContent, err
}

// GetChannelInfo loads the base info for a channel
func (repo ChannelsRepo) GetChannelInfo(slug string) (*ChannelContent, error) {
	channelContentKey := repo.channelContentKeyFor(slug)
	if channelContentKey == nil {
		return nil, errors.New("No channel with slug: " + slug)
	}

	var channelContent = ChannelContent{}
	err := datastore.Get(repo.ctx, channelContentKey, &channelContent)
	channelContent.Key = channelContentKey

	return &channelContent, err
}

// CreatePostInput is used to create new posts
type CreatePostInput struct {
	ChannelSlug          string
	ParentPostKeyEncoded *string
	MarkdownSource       string
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
		return nil, fmt.Errorf("Post content cannot be empty.")
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

// ListPostsInChannel lists all post in a channel of a certain slug
func (repo ChannelsRepo) ListPostsInChannel(channelSlug string) ([]Post, error) {
	channelContentKey := repo.channelContentKeyFor(channelSlug)
	if channelContentKey == nil {
		return nil, errors.New("No channel with slug: " + channelSlug)
	}

	limit := 100
	// q := datastore.NewQuery(postType).Ancestor(channelContentKey).Limit(limit).Filter("ParentPostKey >", nil).Order("ParentPostKey").Order("-CreatedAt")
	// q := datastore.NewQuery(postType).Ancestor(channelContentKey).Limit(limit).Filter("ParentPostKey =", nil).Order("-CreatedAt")
	q := datastore.NewQuery(postType).Ancestor(channelContentKey).Limit(limit).Order("-CreatedAt")
	posts := make([]Post, 0, limit)
	replies := make(map[string][]Post)
	for i := q.Run(repo.ctx); ; {
		var currentPost Post
		key, err := i.Next(&currentPost)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		currentPost.Key = key

		readPostContentFromStorageIfNeeded(repo.ctx, &currentPost)

		if currentPost.ParentPostKey != nil {
			replies[currentPost.ParentPostKey.Encode()] = append(replies[currentPost.ParentPostKey.Encode()], currentPost)
			log.Printf("added reply for %v now %v\n", currentPost.ParentPostKey, replies)
		} else {
			posts = append(posts, currentPost)
		}
	}

	postsWithReplies := make([]Post, 0, len(posts))
	for _, post := range posts {
		postReplies := replies[post.Key.Encode()]
		for i, j := 0, len(postReplies)-1; i < j; i, j = i+1, j-1 {
			postReplies[i], postReplies[j] = postReplies[j], postReplies[i]
		}
		log.Printf("replies for %v are %v\n", post.Key, postReplies)
		post.Replies = &postReplies

		postsWithReplies = append(postsWithReplies, post)
	}

	return postsWithReplies, nil
}
