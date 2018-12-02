package main

import (
	graphql "github.com/graph-gophers/graphql-go"
)

// Actor is implemented by Person
type Actor interface {
	ID() graphql.ID
}

// ActorResolver resolves Actor
type ActorResolver struct {
	Actor
}

// DummyActor used until we have a user system
type DummyActor struct{}

// ID resolved
func (d DummyActor) ID() graphql.ID {
	return graphql.ID("dummy")
}

// // ToPerson converts the receiver to a person, if it is
// func (actor Actor) ToPerson() (*Person, bool) {
// 	person, ok := actor.actor.(Person)
// 	if ok {
// 		return &person, true
// 	}

// 	return nil, false
// }
