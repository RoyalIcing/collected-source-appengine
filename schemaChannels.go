package main

import (
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
