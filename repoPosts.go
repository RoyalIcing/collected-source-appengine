package main

import (
	"context"
	"errors"
	"time"

	"google.golang.org/appengine/datastore"
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

// MediaType resolved
// func (markdownDocument *MarkdownDocument) MediaType() MediaType {
// 	parameters := []string{}
// 	mediaType := NewMediaType("text", "markdown", parameters)
// 	return mediaType
// }

// ChannelContent holds main data of a channel
type ChannelContent struct {
	Key         *datastore.Key `appengine:"-" json:"id"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
}

// ChannelSlug allows a channel to be found by slug
type ChannelSlug struct {
	ContentKey *datastore.Key
}

// Post has a markdown document
type Post struct {
	CreatedAt time.Time      `json:"createdAt"`
	Key       *datastore.Key `appengine:"-" json:"id"`
	//AuthorID string            `json:"authorID"`
	Content MarkdownDocument `json:"content"`
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

// CreatePost creates a new post
func (repo ChannelsRepo) CreatePost(channelSlug string, markdownSource string) (*Post, error) {
	channelContentKey := repo.channelContentKeyFor(channelSlug)
	if channelContentKey == nil {
		return nil, errors.New("No channel with slug: " + channelSlug)
	}

	postKey := datastore.NewIncompleteKey(repo.ctx, postType, channelContentKey)

	markdownDocument := NewMarkdownDocument(markdownSource)
	post := Post{
		Content:   markdownDocument,
		CreatedAt: time.Now().UTC(),
	}
	postKey, err := datastore.Put(repo.ctx, postKey, &post)
	if err != nil {
		return nil, err
	}

	post.Key = postKey

	return &post, nil
}

// GetPostWithIDInChannel lists all post in a channel of a certain slug
func (repo ChannelsRepo) GetPostWithIDInChannel(channelSlug string, id string) (*Post, error) {
	channelContentKey := repo.channelContentKeyFor(channelSlug)
	if channelContentKey == nil {
		return nil, errors.New("No channel with slug: " + channelSlug)
	}

	postKey := datastore.NewKey(repo.ctx, postType, id, 0, channelContentKey)

	var post Post
	err := datastore.Get(repo.ctx, postKey, &post)
	if err == datastore.ErrNoSuchEntity {
		return nil, errors.New("No post with id: " + id)
	}
	if err != nil {
		return nil, errors.New("Error reading post with id: " + id + ": " + err.Error())
	}

	return &post, nil
}

// ListPostsInChannel lists all post in a channel of a certain slug
func (repo ChannelsRepo) ListPostsInChannel(channelSlug string) ([]Post, error) {
	channelContentKey := repo.channelContentKeyFor(channelSlug)
	if channelContentKey == nil {
		return nil, errors.New("No channel with slug: " + channelSlug)
	}

	limit := 100
	q := datastore.NewQuery(postType).Ancestor(channelContentKey).Limit(limit).Order("-__key__")
	posts := make([]Post, 0, limit)
	var currentPost Post
	for i := q.Run(repo.ctx); ; {
		key, err := i.Next(&currentPost)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		currentPost.Key = key
		posts = append(posts, currentPost)
	}
	return posts, nil
}
