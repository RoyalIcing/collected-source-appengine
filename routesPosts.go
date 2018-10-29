package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	// "golang.org/x/net/html"
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

	dynamicElementsEnabled := map[string]bool{"posts": true, "developer": true}

	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts").Methods("GET").
		HandlerFunc(WithHTMLTemplate(listPostsInChannelHTMLHandle, htmlHandlerOptions{dynamicElementsEnabled: dynamicElementsEnabled}))
	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts/{postID}").Methods("GET").
		HandlerFunc(WithHTMLTemplate(showPostInChannelHTMLHandle, htmlHandlerOptions{dynamicElementsEnabled: dynamicElementsEnabled}))
	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts").Methods("POST").
		HandlerFunc(WithHTMLTemplate(createPostInChannelHTMLHandle, htmlHandlerOptions{form: true, dynamicElementsEnabled: dynamicElementsEnabled}))
	r.Path("/org:{orgSlug}/channel:{channelSlug}/posts/{postID}/posts").Methods("POST").
		HandlerFunc(WithHTMLTemplate(createPostInChannelHTMLHandle, htmlHandlerOptions{form: true, dynamicElementsEnabled: dynamicElementsEnabled}))
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

func viewChannelHeader(m ChannelViewModel, fontSize string, w *bufio.Writer) {
	w.WriteString(fmt.Sprintf(`
<header class="pt-4 pb-3 bg-indigo-darker">
	<div class="max-w-md mx-auto flex row justify-between">
		<h1 class="%s">
			<a href="%s" class="text-white no-underline hover:underline">💬 %s</a>
		</h1>
		<input type="search" placeholder="Search %s" class="w-64 px-2 py-2 bg-indigo rounded">
	</div>
</header>
`, fontSize, m.HTMLPostsURL(), m.ChannelSlug, m.ChannelSlug))
}

func htmlError(err error) template.HTML {
	return template.HTML(`<p>` + template.HTMLEscapeString(err.Error()) + `</p>`)
}

func makeViewPostTemplate(ctx context.Context, m ChannelViewModel) *template.Template {
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
		"displayCommandResult": func(post Post) template.HTML {
			if post.CommandType != "" {
				command, err := ParseCommandInput(post.Content.Source)
				if err == nil {
					result, err := command.Run(ctx)
					if err != nil {
						return htmlError(err)
					} else {
						return template.HTML(`<hr class="mt-4 mb-4 border-b border-green">` + SafeHTMLForCommandResult(result))
					}
				} else {
					return htmlError(err)
				}
			}
			return ""
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

{{define "commandResult"}}
{{displayCommandResult .}}
{{end}}

{{define "postActions"}}
<div class="mt-4" data-target="posts.actions">
	<form data-target="posts.createReplyForm" method="post" action="{{childPostsURL .Key.Encode}}" class="my-4"></form>
	<div class="flex row justify-between">
		<div></div>
		<div class="flex row bg-grey-light border border-grey-light rounded">
			<button data-action="posts#beginReply" class="px-2 py-1 text-grey-darkest"> ↩︎</button>
			<button data-action="posts#addToFaves" class="px-2 py-1 text-grey-darkest border-l border-grey-light"> ☆</button>
		</div>
	</div>
</div>
{{end}}

{{define "reply"}}
<div class="pt-4 pb-4 bg-white" data-target="posts.post">
{{template "topBar" .}}
{{template "content" .}}
{{template "postActions" .}}
</div>
{{end}}

{{define "postInList"}}
<div class="p-4 pb-6 bg-white border-b border-grey-dark shadow-md" data-target="posts.post">
{{template "topBar" .}}
{{template "content" .}}
{{template "postActions" .}}

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

<div>
{{template "commandResult" .}}
</div>

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

func viewPostInChannelHTMLHandle(ctx context.Context, post Post, m ChannelViewModel, w *bufio.Writer) {
	t := makeViewPostTemplate(ctx, m)
	t.ExecuteTemplate(w, "postIndividual", post)
}

func viewPostsInChannelHTMLHandle(ctx context.Context, posts []Post, m ChannelViewModel, w *bufio.Writer) {
	t := makeViewPostTemplate(ctx, m)
	for _, post := range posts {
		t.ExecuteTemplate(w, "postInList", post)
	}
}

func viewCreatePostFormInChannelHTMLHandle(channelViewModel ChannelViewModel, w *bufio.Writer) {
	w.WriteString(`
<form data-target="posts.createForm" method="post" action="` + channelViewModel.HTMLPostsURL() + `" class="my-4">
<textarea data-action="input->posts#markdownInputChanged" name="markdownSource" rows="4" placeholder="Write…" class="block w-full p-2 bg-white border border-grey rounded shadow-inner"></textarea>
<div class="flex flex-row-reverse">
<button type="submit" name="action" value="submitPost" data-target="posts.submitPostButton" class="mt-2 px-4 py-2 font-bold text-white bg-blue-darker border border-blue-darker">Post</button>
<button type="submit" name="action" value="runCommand" data-target="posts.runCommandButton" class="mt-2 px-4 py-2 font-bold text-green-dark bg-white border border-green-dark hidden">Run</button>
<button type="submit" name="action" value="beginDraft" data-target="posts.beginDraftButton" class="mt-2 px-4 py-2 font-bold text-white bg-purple-dark border border-purple-dark hidden">Begin Draft</button>
</div>
</form>
`)
}

func viewDeveloperSectionForPostsInChannelHTMLHandle(channelViewModel ChannelViewModel, w *bufio.Writer) {
	query := strings.Replace(`{
	channel(slug: "`+channelViewModel.ChannelSlug+`") {
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
}`, "\t", "  ", -1)

	w.WriteString(`
<div data-controller="developer">
<details class="mb-4 bg-indigo-darker">
	<summary class="max-w-md mx-auto p-1 italic cursor-pointer text-right text-sm text-indigo-lighter bg-indigo-darker select-none">Developer</summary>
	<div class="max-h-screen overflow-auto bg-yellow-lightest">
		<div class="flex row">
			<pre class="w-1/2 p-2 bg-indigo-lightest text-indigo-darkest"><code data-target="developer.queryCode">` + query + `</code></pre>
			<div class="w-1/2">
				<button data-action="developer#runQuery" class="mb-2 px-4 py-2 font-bold text-white bg-green-darker border border-green-darker">Query</button>
				<pre class="p-2 text-green-darkest"><code data-target="developer.result"></code></pre>
			</div>
		</div>
	</div>
</details>
</div>
`)
}

func viewPostsInChannelHTMLPartial(ctx context.Context, errs []error, channelViewModel ChannelViewModel, posts []Post, viewSection func(wide bool, viewInner func(sw *bufio.Writer))) {
	viewSection(true, func(sw *bufio.Writer) {
		viewChannelHeader(channelViewModel, "text-2xl text-center", sw)
		viewDeveloperSectionForPostsInChannelHTMLHandle(channelViewModel, sw)
	})

	viewSection(false, func(sw *bufio.Writer) {
		for _, err := range errs {
			viewErrorMessage(err.Error(), sw)
		}

		sw.WriteString(`<div data-controller="posts">`)
		viewCreatePostFormInChannelHTMLHandle(channelViewModel, sw)
		viewPostsInChannelHTMLHandle(ctx, posts, channelViewModel, sw)
		sw.WriteString(`</div>`)
	})
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

	channelViewModel := vars.ToChannelViewModel()
	channelViewModel.Org.ViewPage(w, func(viewSection func(wide bool, viewInner func(sw *bufio.Writer))) {
		if err != nil {
			viewSection(false, func(sw *bufio.Writer) {
				viewErrorMessage("Error listing posts: "+err.Error(), sw)
			})
			return
		}
		viewPostsInChannelHTMLPartial(ctx, nil, channelViewModel, posts, viewSection)
	})
}

func showPostInChannelHTMLHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	post, err := channelsRepo.GetPostWithIDInChannel(vars.channelSlug(), vars.postID())
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "Error loading post: "+err.Error())
		return
	}

	w.WriteHeader(200)

	channelViewModel := vars.ToChannelViewModel()
	channelViewModel.Org.ViewPage(w, func(viewSection func(wide bool, viewInner func(sw *bufio.Writer))) {
		viewSection(true, func(sw *bufio.Writer) {
			viewChannelHeader(channelViewModel, "text-2xl text-center", sw)
		})

		viewSection(false, func(sw *bufio.Writer) {
			sw.WriteString(`<div data-controller="posts">`)
			viewPostInChannelHTMLHandle(ctx, *post, channelViewModel, sw)

			sw.WriteString(`<div class="hidden">`)
			viewCreatePostFormInChannelHTMLHandle(channelViewModel, sw)
			sw.WriteString(`</div>`)

			sw.WriteString(`</div>`)
		})
	})
}

func createPostInChannelHTMLHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	orgRepo := NewOrgRepo(ctx, vars.orgSlug())
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	action := r.PostFormValue("action")

	commandType := ""
	if action == "runCommand" {
		commandType = "v0"
	}

	input := CreatePostInput{
		ChannelSlug:          vars.channelSlug(),
		ParentPostKeyEncoded: vars.optionalPostID(),
		MarkdownSource:       r.PostFormValue("markdownSource"),
		CommandType:          commandType,
	}
	_, errCreating := channelsRepo.CreatePost(input)

	posts, errListing := channelsRepo.ListPostsInChannel(vars.channelSlug())

	channelViewModel := vars.ToChannelViewModel()
	channelViewModel.Org.ViewPage(w, func(viewSection func(wide bool, viewInner func(sw *bufio.Writer))) {
		var errs []error
		if errCreating != nil {
			errs = append(errs, fmt.Errorf("Error creating post: %s", errCreating.Error()))
		}
		if errListing != nil {
			errs = append(errs, fmt.Errorf("Error listing posts"))
		}
		viewPostsInChannelHTMLPartial(ctx, errs, channelViewModel, posts, viewSection)
	})
}
