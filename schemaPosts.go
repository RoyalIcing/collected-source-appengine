package main

import (
	graphql "github.com/graph-gophers/graphql-go"
)

// Author resolved
// func (post *Post) Author() *Actor {
// 	person := NewPerson("999", "Example", "Yep", []string{}, map[string]string{})
// 	actor := Actor{actor: person}
// 	return &actor
// }

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
func (postEdge *PostEdge) Node() *Post {
	return postEdge.post
}

// Cursor resolved
func (postEdge *PostEdge) Cursor() graphql.ID {
	return graphql.ID(postEdge.cursor)
}

// PostConnection is a connection to a collection of posts
type PostConnection struct {
	edges *[]*PostEdge
}

// NewPostConnection makes a post connection with the provided values
func NewPostConnection(edges *[]*PostEdge) PostConnection {
	postConnection := PostConnection{
		edges: edges,
	}
	return postConnection
}

// Edges resolved
func (postConnection PostConnection) Edges() *[]*PostEdge {
	return postConnection.edges
}

// TotalCount resolved
func (postConnection PostConnection) TotalCount() *int32 {
	if postConnection.edges == nil {
		return nil
	}

	count := int32(len(*postConnection.edges))
	return &count
}
