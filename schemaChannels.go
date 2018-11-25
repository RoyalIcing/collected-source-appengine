package main

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
)

// Channel is a channel
type Channel struct {
	slug string
}

// NewChannel makes a Channel
func NewChannel(slug string) *Channel {
	channel := Channel{
		slug: slug,
	}
	return &channel
}

// ID resolved
func (channel *Channel) ID() graphql.ID {
	return graphql.ID(channel.slug)
}

// Slug resolved
func (channel *Channel) Slug() *string {
	return &channel.slug
}

// Posts resolved
func (channel *Channel) Posts(ctx context.Context) (*PostsConnection2, error) {
	orgSlug := "RoyalIcing"

	orgRepo := NewOrgRepo(ctx, orgSlug)
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	posts, err := channelsRepo.ListPostsInChannel(channel.slug)
	if err != nil {
		return nil, fmt.Errorf("Error loading posts: %s", err.Error())
	}

	postEdges := make([]*PostEdge, 0, len(posts))
	for _, post := range posts {
		localPost := post
		postEdge := NewPostEdge(&localPost, "-")
		postEdges = append(postEdges, postEdge)
	}

	c := NewPostsConnectionWithEdges(&postEdges)

	return &c, nil
}
