package main

import (
	"encoding/csv"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

// AddAPIPostsRoutes adds routes for working with channels and posts with JSON/CSV
func AddAPIPostsRoutes(r *mux.Router) {
	// TODO: move to separate routesChannel.go
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}").Methods("GET").
		HandlerFunc(getChannelInfoHandle)
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}").Methods("PUT").
		HandlerFunc(createChannelHandle)

	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts").Methods("GET").
		HandlerFunc(listPostsInChannelHandle)
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts.csv").Methods("GET").
		HandlerFunc(listPostsCSVInChannelHandle)
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts/{postID}").Methods("GET").
		HandlerFunc(getPostInChannelHandle)
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts").Methods("POST").
		HandlerFunc(createPostInChannelHandle)
}

func getChannelInfoHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	channel, err := channelsRepo.GetChannelInfo(vars.channelSlug())
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, channel)
}

func createChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	channel, err := channelsRepo.CreateChannel(vars.channelSlug())
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, channel)
}

func listPostsInChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	postsConnection := channelsRepo.NewPostsConnection(PostsConnectionOptions{
		channelSlug:    vars.channelSlug(),
		includeReplies: true,
		maxCount:       1000,
	})

	w.Header().Add("Content-Type", "text/json")

	posts, err := postsConnection.All()
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, posts)
}

func listPostsCSVInChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	postsConnection := channelsRepo.NewPostsConnection(PostsConnectionOptions{
		channelSlug:    vars.channelSlug(),
		includeReplies: false,
		maxCount:       1000,
	})

	w.Header().Add("Content-Type", "text/csv")

	csvWriter := csv.NewWriter(w)
	err := postsConnection.WriteToCSV(csvWriter)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	csvWriter.Flush()
}

func getPostInChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	posts, err := channelsRepo.GetPostWithIDInChannel(vars.channelSlug(), vars.postID())
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, posts)
}

type createPostBody struct {
	MarkdownSource string `json:"markdownSource"`
}

func createPostInChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	bodyDecoder := json.NewDecoder(r.Body)
	var body createPostBody
	err := bodyDecoder.Decode(&body)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	input := CreatePostInput{
		ChannelSlug:          vars.channelSlug(),
		ParentPostKeyEncoded: nil,
		MarkdownSource:       body.MarkdownSource,
	}

	post, err := channelsRepo.CreatePost(input)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, post)
}
