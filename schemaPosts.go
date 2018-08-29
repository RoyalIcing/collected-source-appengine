package main

import (
	graphql "github.com/graph-gophers/graphql-go"
)

// MarkdownDocumentResolver decorates a MarkdownDocument for GraphQL
type MarkdownDocumentResolver struct {
	MarkdownDocument
}

// Source resolved
func (markdownDocument *MarkdownDocument) source() string {
	return markdownDocument.Source
}

// Source resolved
func (r *MarkdownDocumentResolver) Source() *string {
	s := r.source()
	return &s
}

// MediaType resolved
func (r *MarkdownDocumentResolver) MediaType() MediaType {
	parameters := []string{}
	mediaType := NewMediaType("text", "markdown", parameters)
	return mediaType
}

// PostResolver decorates a Post for GraphQL
type PostResolver struct {
	Post
}

// ID resolved
func (r *PostResolver) ID() graphql.ID {
	return graphql.ID(r.Key.Encode())
}

func (post Post) content() MarkdownDocument {
	return post.Content
}

// Content resolved
func (r *PostResolver) Content() *MarkdownDocumentResolver {
	return &MarkdownDocumentResolver{r.content()}
}

// Author resolved
func (post *Post) Author() *ActorResolver {
	dummy := DummyActor{}
	actor := ActorResolver{dummy}
	return &actor
}

// PostEdge is a reference to a post within a connection
type PostEdge struct {
	post   *Post
	cursor string
}

// NewPostEdge makes a post edge with the provided values
func NewPostEdge(post *Post, cursor string) *PostEdge {
	postEdge := PostEdge{
		post:   post,
		cursor: cursor,
	}
	return &postEdge
}

// Node resolved
func (postEdge *PostEdge) Node() *PostResolver {
	if postEdge.post == nil {
		return nil
	}
	return &PostResolver{*postEdge.post}
}

// Cursor resolved
func (postEdge *PostEdge) Cursor() graphql.ID {
	return graphql.ID(postEdge.cursor)
}

// PostsConnection is a connection to a collection of posts
type PostsConnection struct {
	edges *[]*PostEdge
}

// NewPostsConnection makes a post connection with the provided values
func NewPostsConnection(edges *[]*PostEdge) PostsConnection {
	postsConnection := PostsConnection{
		edges: edges,
	}
	return postsConnection
}

// Edges resolved
func (postsConnection PostsConnection) Edges() *[]*PostEdge {
	return postsConnection.edges
}

// TotalCount resolved
func (postsConnection PostsConnection) TotalCount() *int32 {
	if postsConnection.edges == nil {
		return nil
	}

	count := int32(len(*postsConnection.edges))
	return &count
}
