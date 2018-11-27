package main

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"

	"google.golang.org/appengine/datastore"
)

const (
	channelSlugType    = "ChannelSlug"
	channelContentType = "ChannelContent"
)

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
func (repo ChannelsRepo) CreateChannel(slug string) (*ChannelContent, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("Channel slug cannot be empty")
	}

	channelSlugKey := repo.channelSlugKeyFor(slug)

	channelContentKey := datastore.NewIncompleteKey(repo.ctx, channelContentType, repo.orgRepo.RootKey())
	channelContent := ChannelContent{
		Slug:        slug,
		Description: "",
	}
	channelContentKey, err := datastore.Put(repo.ctx, channelContentKey, &channelContent)

	channelSlug := ChannelSlug{
		ContentKey: channelContentKey,
	}
	// FIXME: should create channel slug first? As this is the whole point of having the channel slug type, to unique the slugs
	// FIXME: needs to use datastore.RunInTransaction
	_, err = datastore.Put(repo.ctx, channelSlugKey, &channelSlug)
	if err != nil {
		return nil, err
	}

	channelContent.Key = channelContentKey
	return &channelContent, nil
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

// OrgChannelsConnectionOptions offers parameters when retrieving channels from an org
type OrgChannelsConnectionOptions struct {
	maxCount int
}

// OrgChannelsConnection allow retrieving channels from an org
type OrgChannelsConnection struct {
	ctx     context.Context
	orgKey  *datastore.Key
	options OrgChannelsConnectionOptions
}

// NewChannelsConnection allows enumerating through the channels of an org
func (repo ChannelsRepo) NewChannelsConnection(options OrgChannelsConnectionOptions) *OrgChannelsConnection {
	c := OrgChannelsConnection{ctx: repo.ctx, orgKey: repo.orgRepo.RootKey(), options: options}
	return &c
}

// Enumerate loops through each channel
func (c *OrgChannelsConnection) Enumerate(useChannel func(channel ChannelContent)) error {
	ctx := c.ctx
	orgKey := c.orgKey
	limit := c.options.maxCount

	q := datastore.NewQuery(channelContentType).Ancestor(orgKey).Limit(limit)
	for i := q.Run(ctx); ; {
		var currentChannel ChannelContent
		key, err := i.Next(&currentChannel)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}

		currentChannel.Key = key

		useChannel(currentChannel)
	}

	return nil
}

// All gets all the channels as a slice
func (c *OrgChannelsConnection) All() ([]ChannelContent, error) {
	var channels []ChannelContent
	err := c.Enumerate(func(channel ChannelContent) {
		channels = append(channels, channel)
	})
	return channels, err
}

// WriteToCSV writes all the channels as CSV records
func (c *OrgChannelsConnection) WriteToCSV(w *csv.Writer) error {
	w.Write([]string{"id", "slug", "description"})

	return c.Enumerate(func(channel ChannelContent) {
		w.Write([]string{channel.Key.Encode(), channel.Slug, channel.Description})
	})
}
