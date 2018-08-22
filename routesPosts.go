package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"time"
	// "time"

	"html/template"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

// AddPostsRoutes adds routes for working with channels and posts
func AddPostsRoutes(r *mux.Router) {
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}").Methods("GET").
		HandlerFunc(getChannelInfoHandle)
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}").Methods("PUT").
		HandlerFunc(createChannelHandle)

	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts").Methods("GET").
		HandlerFunc(listPostsInChannelHandle)
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts/{postID}").Methods("GET").
		HandlerFunc(getPostInChannelHandle)
	r.Path("/1/org:{orgSlug}/channel:{channelSlug}/posts").Methods("POST").
		HandlerFunc(createPostInChannelHandle)

	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts").Methods("GET").
		HandlerFunc(withHTMLTemplate(listPostsInChannelHTMLHandle))
	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts").Methods("POST").
		HandlerFunc(withHTMLTemplate(createPostInChannelHTMLHandle))

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

	posts, err := channelsRepo.ListPostsInChannel(vars.channelSlug())
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, posts)
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

	post, err := channelsRepo.CreatePost(vars.channelSlug(), body.MarkdownSource)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, post)
}

func withHTMLTemplate(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		header.Set("X-Content-Type-Options", "nosniff")

		io.WriteString(w, `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<link href="https://cdn.jsdelivr.net/npm/tailwindcss/dist/tailwind.min.css" rel="stylesheet">
</head>
<body class="bg-blue-light">
<main class="max-w-md mx-auto mt-4 mb-8">
`)

		f(w, r)

		io.WriteString(w, "</main></body></html>")
	})
}

func viewPostsInChannelHTMLHandle(posts []Post, w *bufio.Writer) {
	t := template.Must(template.New("createdAt").Parse(`
<div class="p-4 bg-white border-b border-blue-light">
<div>
<span class="font-bold">Name</span>
<span class="text-grey-dark">@handle</span>
·
<time datetime="{{.CreatedAtRFC3339}}">{{.CreatedAtDisplay}}</time>
</div>
<p class="whitespace-pre-wrap">
{{.MarkdownSource}}
</p>
</div>
`))

	for _, post := range posts {
		t.Execute(w, struct {
			CreatedAtDisplay string
			CreatedAtRFC3339 string
			MarkdownSource   string
		}{
			CreatedAtDisplay: post.CreatedAt.Format(time.RFC822),
			CreatedAtRFC3339: post.CreatedAt.Format(time.RFC3339),
			MarkdownSource:   post.Content.Source,
		})
	}
}

func viewCreatePostFormInChannelHTMLHandle(vars RouteVars, w *bufio.Writer) {
	w.WriteString(`
<form method="post" action="/org:` + vars.orgSlug() + `/channel:` + vars.channelSlug() + `/posts" class="my-4">
<textarea name="markdownSource" rows="4" class="block w-full p-2 border border-blue"></textarea>
<button type="submit" class="mt-2 px-4 py-2 text-white bg-blue-darkest">Post</button>
</form>
`)
}

func listPostsInChannelHTMLHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	posts, err := channelsRepo.ListPostsInChannel(vars.channelSlug())
	if err != nil {
		io.WriteString(w, "Error: "+err.Error())
		return
	}

	sw := bufio.NewWriter(w)
	defer sw.Flush()

	w.WriteHeader(200)

	sw.WriteString("<h1>Posts</h1>")
	viewCreatePostFormInChannelHTMLHandle(vars, sw)
	viewPostsInChannelHTMLHandle(posts, sw)
}

func createPostInChannelHTMLHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, "Invalid form request: "+err.Error())
		return
	}

	markdownSource := r.PostFormValue("markdownSource")

	_, err = channelsRepo.CreatePost(vars.channelSlug(), markdownSource)
	if err != nil {
		io.WriteString(w, "Error creating post: "+err.Error())
		return
	}

	posts, err := channelsRepo.ListPostsInChannel(vars.channelSlug())
	if err != nil {
		io.WriteString(w, "Error listing posts: "+err.Error())
		return
	}

	sw := bufio.NewWriter(w)
	defer sw.Flush()

	sw.WriteString("<h1>Posts</h1>")
	viewCreatePostFormInChannelHTMLHandle(vars, sw)
	viewPostsInChannelHTMLHandle(posts, sw)
}
