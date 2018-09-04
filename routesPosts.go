package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
		HandlerFunc(withHTMLTemplate(listPostsInChannelHTMLHandle, htmlHandlerOptions{}))
	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts/{postID}").Methods("GET").
		HandlerFunc(withHTMLTemplate(showPostInChannelHTMLHandle, htmlHandlerOptions{}))
	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts").Methods("POST").
		HandlerFunc(withHTMLTemplate(createPostInChannelHTMLHandle, htmlHandlerOptions{form: true}))
	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts/{postID}/posts").Methods("POST").
		HandlerFunc(withHTMLTemplate(createPostInChannelHTMLHandle, htmlHandlerOptions{form: true}))
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

type htmlHandlerOptions struct {
	form bool
}

func withHTMLTemplate(f http.HandlerFunc, options htmlHandlerOptions) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		header.Set("X-Content-Type-Options", "nosniff")

		var formErr error
		if options.form {
			formErr = r.ParseForm()
		}

		io.WriteString(w, `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<link href="https://cdn.jsdelivr.net/npm/tailwindcss/dist/tailwind.min.css" rel="stylesheet">
<script defer src="https://unpkg.com/stimulus@1.0.1/dist/stimulus.umd.js"></script>
</head>
<body class="bg-blue-light">
`)

		if formErr != nil {
			w.WriteHeader(400)
			io.WriteString(w, "Invalid form request: "+formErr.Error())
		} else {
			f(w, r)
		}

		io.WriteString(w, `
<script>
document.addEventListener("DOMContentLoaded", () => {
const app = Stimulus.Application.start();

app.register('posts', class extends Stimulus.Controller {
  static get targets() {
		return [ 'post', 'replyHolder' ];
	}

  beginReply({ target: button }) {
		const post = button.parentElement;
		const createReplyForm = post.querySelector('[data-target="posts.createReplyForm"]');
		const createForm = this.targets.find('createForm'); // this.createFormTarget;
		createReplyForm.innerHTML = createForm.innerHTML;
    //this.replyFieldTarget.textContent = "This is my reply"
  }
});
});
</script>
`)

		io.WriteString(w, "</body></html>")
	})
}

func viewErrorMessage(errorMessage string, w *bufio.Writer) {
	w.WriteString(`<p class="py-1 px-2 bg-white text-red">` + errorMessage + "</p>")
}

func viewChannelHeader(m ChannelViewModel, fontSize string, w *bufio.Writer) {
	w.WriteString(fmt.Sprintf(`
<h1 class="%s mb-4">
<a href="%s" class="text-black no-underline hover:underline">💬 %s</a>
</h1>
`, fontSize, m.HTMLPostsURL(), m.ChannelSlug))
}

func makeViewPostTemplate(m ChannelViewModel) *template.Template {
	t := template.New("post").Funcs(template.FuncMap{
		"postURL": func(postID string) string {
			return m.HTMLPostURL(postID)
		},
		"childPostsURL": func(postID string) string {
			return m.HTMLPostChildPostsURL(postID)
		},
		"formatMarkdown": func(markdownSource string) string {
			return strings.TrimSpace(markdownSource)
		},
		"formatTimeRFC3339": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"displayTime": func(t time.Time) string {
			return t.Format(time.RFC822)
		},
	})
	t = template.Must(t.Parse(`
{{define "topBar"}}
<div>
<span class="font-bold">Name</span>
<span class="text-grey-dark">@handle</span>
·
<a href="{{postURL .Key.Encode}}" class="text-grey-dark no-underline hover:underline">
<time datetime="{{formatTimeRFC3339 .CreatedAt}}">{{displayTime .CreatedAt}}</time>
</a>
</div>
{{end}}

{{define "topBarLarge"}}
<div>
<span class="text-lg font-bold">Name</span>
<span class="text-grey-dark">@handle</span>
·
<a href="{{postURL .Key.Encode}}" class="text-grey-dark no-underline hover:underline">
<time datetime="{{formatTimeRFC3339 .CreatedAt}}">{{displayTime .CreatedAt}}</time>
</a>
</div>
{{end}}

{{define "content"}}
<p class="whitespace-pre-wrap">
{{formatMarkdown .Content.Source}}
</p>
{{end}}

{{define "contentLarge"}}
<p class="text-xl whitespace-pre-wrap">
{{formatMarkdown .Content.Source}}
</p>
{{end}}

{{define "reply"}}
<div class="pt-4 pb-4 bg-white" data-target="posts.post">
{{template "topBar" .}}
{{template "content" .}}
</div>
{{end}}

{{define "postInList"}}
<div class="p-4 pb-6 bg-white border-b border-blue-light" data-target="posts.post">
{{template "topBar" .}}
{{template "content" .}}

<div class="mt-4">
	<form data-target="posts.createReplyForm" method="post" action="{{childPostsURL .Key.Encode}}" class="my-4"></form>
	<button data-action="posts#beginReply" class="px-2 py-1 bg-grey-lighter"> ↩︎</button>
</div>

<div data-target="posts.replies">
{{range .Replies}}
{{template "reply" .}}
{{end}}
</div>
</div>
{{end}}

{{define "postIndividual"}}
<div class="p-4 pb-6 bg-white border-b border-blue-light" data-target="posts.post">
{{template "topBarLarge" .}}
{{template "contentLarge" .}}

<div class="mt-4">
	<form data-target="posts.createReplyForm" method="post" action="{{childPostsURL .Key.Encode}}" class="my-4"></form>
	<button data-action="posts#beginReply" class="px-2 py-1 bg-grey-lighter"> ↩︎</button>
</div>

<div data-target="posts.replies">
{{range .Replies}}
{{template "reply" .}}
{{end}}
</div>
</div>
{{end}}
`))

	return t
}

func viewPostInChannelHTMLHandle(post Post, m ChannelViewModel, w *bufio.Writer) {
	t := makeViewPostTemplate(m)
	t.ExecuteTemplate(w, "postIndividual", post)
}

func viewPostsInChannelHTMLHandle(posts []Post, m ChannelViewModel, w *bufio.Writer) {
	t := makeViewPostTemplate(m)
	for _, post := range posts {
		t.ExecuteTemplate(w, "postInList", post)
	}
}

func viewCreatePostFormInChannelHTMLHandle(vars RouteVars, w *bufio.Writer) {
	w.WriteString(`
<form data-target="posts.createForm" method="post" action="/org:` + vars.orgSlug() + `/channel:` + vars.channelSlug() + `/posts" class="my-4">
<textarea name="markdownSource" rows="4" placeholder="Write…" class="block w-full p-2 border border-blue rounded"></textarea>
<button type="submit" class="mt-2 px-4 py-2 text-white bg-blue-darkest">Post</button>
</form>
`)
}

func viewDeveloperSectionForPostsInChannelHTMLHandle(vars RouteVars, w *bufio.Writer) {
	w.WriteString(`
<details class="mb-4">
<summary class="italic cursor-pointer text-sm">Developer</summary>
<pre class="mt-2 p-2 bg-black text-white"><code>{
  channel(slug: "` + vars.channelSlug() + `") {
    slug
    posts {
      totalCount
      edges {
        node {
          id
          content {
            source
            mediaType {
              baseType
              subtype
            }
          }
        }
      }
    }
  }
}</code></pre>
</details>
`)
}

func listPostsInChannelHTMLHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	posts, err := channelsRepo.ListPostsInChannel(vars.channelSlug())
	if err != nil {
		io.WriteString(w, "Error loading posts: "+err.Error())
		return
	}

	w.WriteHeader(200)

	channelViewModel := vars.ToChannelViewModel()
	channelViewModel.Org.ViewPage(w, func(sw *bufio.Writer) {
		viewChannelHeader(channelViewModel, "text-4xl text-center", sw)
		viewDeveloperSectionForPostsInChannelHTMLHandle(vars, sw)
		sw.WriteString(`<div data-controller="posts">`)
		viewCreatePostFormInChannelHTMLHandle(vars, sw)
		viewPostsInChannelHTMLHandle(posts, channelViewModel, sw)
		sw.WriteString(`</div>`)
	})
}

func showPostInChannelHTMLHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	post, err := channelsRepo.GetPostWithIDInChannel(vars.channelSlug(), vars.postID())
	if err != nil {
		io.WriteString(w, "Error loading post: "+err.Error())
		return
	}

	w.WriteHeader(200)

	channelViewModel := vars.ToChannelViewModel()
	channelViewModel.Org.ViewPage(w, func(sw *bufio.Writer) {
		viewChannelHeader(channelViewModel, "text-2xl text-center", sw)
		sw.WriteString(`<div data-controller="posts">`)
		viewPostInChannelHTMLHandle(*post, channelViewModel, sw)
		sw.WriteString(`</div>`)
	})
}

func createPostInChannelHTMLHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	input := CreatePostInput{
		ChannelSlug:          vars.channelSlug(),
		ParentPostKeyEncoded: vars.optionalPostID(),
		MarkdownSource:       r.PostFormValue("markdownSource"),
	}

	channelViewModel := vars.ToChannelViewModel()
	channelViewModel.Org.ViewPage(w, func(sw *bufio.Writer) {
		viewChannelHeader(channelViewModel, "text-4xl text-center", sw)
		viewDeveloperSectionForPostsInChannelHTMLHandle(vars, sw)

		defer viewCreatePostFormInChannelHTMLHandle(vars, sw)

		_, err := channelsRepo.CreatePost(input)
		if err != nil {
			viewErrorMessage("Error creating post: "+err.Error(), sw)
			return
		}

		posts, err := channelsRepo.ListPostsInChannel(vars.channelSlug())
		if err != nil {
			viewErrorMessage("Error listing posts: "+err.Error(), sw)
			return
		}

		defer viewPostsInChannelHTMLHandle(posts, channelViewModel, sw)
	})
}
