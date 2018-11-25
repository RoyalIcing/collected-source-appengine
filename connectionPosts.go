package main

import (
	"encoding/csv"
	"errors"
	"log"

	"google.golang.org/appengine/datastore"
)

type PostsConnectionOptions struct {
	channelSlug    string
	includeReplies bool
	maxCount       int
}

type PostsConnection struct {
	repo    ChannelsRepo
	options PostsConnectionOptions
}

func (c *PostsConnection) enumerate(usePost func(post Post)) error {
	ctx := c.repo.ctx
	channelSlug := c.options.channelSlug
	includeReplies := c.options.includeReplies
	limit := c.options.maxCount

	channelContentKey := c.repo.channelContentKeyFor(channelSlug)
	if channelContentKey == nil {
		return errors.New("No channel with slug: " + channelSlug)
	}

	// q := datastore.NewQuery(postType).Ancestor(channelContentKey).Limit(limit).Filter("ParentPostKey >", nil).Order("ParentPostKey").Order("-CreatedAt")
	// q := datastore.NewQuery(postType).Ancestor(channelContentKey).Limit(limit).Filter("ParentPostKey =", nil).Order("-CreatedAt")
	q := datastore.NewQuery(postType).Ancestor(channelContentKey).Limit(limit).Order("-CreatedAt")
	posts := make([]Post, 0, limit)
	replies := make(map[string][]Post)
	for i := q.Run(ctx); ; {
		var currentPost Post
		key, err := i.Next(&currentPost)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}

		currentPost.Key = key

		readPostContentFromStorageIfNeeded(ctx, &currentPost)

		if includeReplies {
			if currentPost.ParentPostKey != nil {
				replies[currentPost.ParentPostKey.Encode()] = append(replies[currentPost.ParentPostKey.Encode()], currentPost)
				log.Printf("added reply for %v now %v\n", currentPost.ParentPostKey, replies)
			} else {
				posts = append(posts, currentPost)
			}
		} else {
			usePost(currentPost)
		}
	}

	if !includeReplies {
		return nil
	}

	for _, post := range posts {
		postReplies := replies[post.Key.Encode()]
		for i, j := 0, len(postReplies)-1; i < j; i, j = i+1, j-1 {
			postReplies[i], postReplies[j] = postReplies[j], postReplies[i]
		}
		log.Printf("replies for %v are %v\n", post.Key, postReplies)
		post.Replies = &postReplies

		usePost(post)
	}

	return nil
}

// All gets all the posts as a slice
func (c *PostsConnection) All() ([]Post, error) {
	var posts []Post
	err := c.enumerate(func(post Post) {
		posts = append(posts, post)
	})
	return posts, err
}

// WriteToCSV writes all the posts as CSV records
func (c *PostsConnection) WriteToCSV(w *csv.Writer) error {
	w.Write([]string{"id", "createdAt", "parentPostID", "commandType", "content"})
	log.Println("wrote csv header")

	return c.enumerate(func(post Post) {
		parentPostID := ""
		if post.ParentPostKey != nil {
			parentPostID = post.ParentPostKey.Encode()
		}
		w.Write([]string{post.Key.Encode(), post.CreatedAt.String(), parentPostID, post.CommandType, post.Content.Source})
		log.Println("wrote csv row")
	})
}
