package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

// AddFeedPostsRoutes adds routes for posts’ RSS/Atom feeds
func AddFeedPostsRoutes(r *mux.Router) {
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts.rss").Methods("GET").
		HandlerFunc(listPostsRSSInChannelHandle)
}

type postsFeedURLMaker struct {
	baseURL     string
	orgSlug     string
	channelSlug string
}

func (m *postsFeedURLMaker) url() string {
	return m.baseURL + "/org:" + m.orgSlug + "/channel:" + m.channelSlug + "/posts"
}

func (m *postsFeedURLMaker) itemURL(id string) string {
	return m.url() + "/" + id
}

func schemeForRequest(r *http.Request) string {
	if IsDev() {
		return "http"
	}

	return "https"
}

func listPostsRSSInChannelHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	urlMaker := &postsFeedURLMaker{
		baseURL:     schemeForRequest(r) + "://" + r.Host,
		orgSlug:     vars.orgSlug(),
		channelSlug: vars.channelSlug(),
	}

	postsConnection := channelsRepo.NewPostsConnection(PostsConnectionOptions{
		channelSlug:    vars.channelSlug(),
		includeReplies: false,
		maxCount:       1000,
	})

	WriteConnectionRSSFeedToHTTP(postsConnection, urlMaker, w)
}
