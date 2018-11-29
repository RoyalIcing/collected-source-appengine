package main

import (
	"encoding/csv"
	"io"
	"net/http"

	"github.com/gorilla/feeds"
)

// Connection represents a collection of items which can be enumerated through
type Connection interface {
	WriteToCSV(w csv.Writer) error
}

type FeedURLMaker interface {
	url() string
	itemURL(id string) string
}

// ConnectionWithFeed represents a collection of items which generate a RSS/Atom feed
type ConnectionWithFeed interface {
	MakeFeed(urlMaker FeedURLMaker) (*feeds.Feed, error)
}

// WriteConnectionRSSFeedToHTTP generates an RSS feed from the connection and writes it to the HTTP response
func WriteConnectionRSSFeedToHTTP(c ConnectionWithFeed, urlMaker FeedURLMaker, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/rss+xml")

	feed, err := c.MakeFeed(urlMaker)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	rssString, err := feed.ToRss()
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	io.WriteString(w, rssString)
}
