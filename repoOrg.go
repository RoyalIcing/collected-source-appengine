package main

import (
	"context"
	// "encoding/json"
	// "net/http"
	// "time"

	// "google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
)

// Org represents a group of people working together
type Org struct {
	Slug string `json:"slug"`
}

// OrgRepo lets you query the channels repository
type OrgRepo struct {
	ctx     context.Context
	orgKey  *datastore.Key
	orgSlug string
}

// NewOrgRepo makes a new org repository with the given org slug
func NewOrgRepo(ctx context.Context, orgSlug string) OrgRepo {
	orgKey := datastore.NewKey(ctx, "Org", orgSlug, 0, nil)
	return OrgRepo{
		ctx:     ctx,
		orgKey:  orgKey,
		orgSlug: orgSlug,
	}
}

// CreateOrg creates the org if necessary
func (repo OrgRepo) CreateOrg() (Org, error) {
	org := Org{Slug: repo.orgSlug}
	_, err := datastore.Put(repo.ctx, repo.orgKey, &org)
	return org, err
}

// RootKey returns the root key for queries
func (repo OrgRepo) RootKey() *datastore.Key {
	return repo.orgKey
}
