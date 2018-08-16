package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

// Adds routes for working with channels and posts
func AddPostsRoutes(r *mux.Router) {
	r.Path("/1/org:{orgName}/channel:{channelName}").Methods("GET").
		HandlerFunc(getChannelInfoHandle)
	r.Path("/1/org:{orgName}/channel:{channelName}").Methods("PUT").
		HandlerFunc(createChannelHandle)

	r.Path("/1/org:{orgName}/channel:{channelName}/posts").Methods("GET").
		HandlerFunc(listPostsInChannelHandle)
	r.Path("/1/org:{orgName}/channel:{channelName}/posts").Methods("POST").
		HandlerFunc(createPostInChannelHandle)
}

func getChannelInfoHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	orgRepo := NewOrgRepo(ctx, "RoyalIcing")
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	channel, err := channelsRepo.GetChannelInfo("design")
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, channel)
}

func createChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	orgRepo := NewOrgRepo(ctx, "RoyalIcing")
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	channel, err := channelsRepo.CreateChannel("design")
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, channel)
}

func listPostsInChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	orgRepo := NewOrgRepo(ctx, "RoyalIcing")
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	posts, err := channelsRepo.ListPostsInChannel("design")
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, posts)
}

func createPostInChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	orgRepo := NewOrgRepo(ctx, "RoyalIcing")
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	post, err := channelsRepo.CreatePost("design", "# Hello")
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, post)
}
