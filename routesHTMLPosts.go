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

	"golang.org/x/net/html"
	"html/template"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

// AddHTMLPostsRoutes adds user-facing routes for channel posts
func AddHTMLPostsRoutes(r *mux.Router) {
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

func viewChannelHeader(m ChannelViewModel, fontSize string, w *bufio.Writer) {
	w.WriteString(fmt.Sprintf(`
<header class="pt-4 pb-3 bg-indigo-darker">
	<div class="max-w-md mx-auto">
		<div class="mx-2 md:mx-0 flex flex-wrap flex-col sm:flex-row items-center sm:items-start sm:justify-between">
			<h1 class="%s min-w-full sm:min-w-0 mb-2 sm:mb-0">
				<a href="%s" class="text-white no-underline hover:underline">💬 %s</a>
			</h1>
			<input type="search" placeholder="Search %s" class="w-64 px-2 py-2 bg-indigo rounded">
		</div>
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
			commandType := post.CommandType
			if commandType != "" {
				if commandType == "v0-runGraphQLQuery" {
					resolver := NewDataStoreResolver()
					schema := MakeSchema(&resolver)
					response := schema.Exec(ctx, post.Content.Source, "", map[string]interface{}{})
					responseJSON, err := json.MarshalIndent(response, "", "  ")
					if err != nil {
						return htmlError(err)
					} else {
						return template.HTML(`<div class="p-2 border-t-2 border-purple bg-purple-lightest rounded-sm"><pre>` + html.EscapeString(string(responseJSON)) + `</pre></div>`)
					}
				} else {
					command, err := ParseCommandInput(post.Content.Source)
					if err == nil {
						result, err := command.Run(ctx)
						if err != nil {
							return htmlError(err)
						} else {
							return template.HTML(`<div class="p-2 border-t-2 border-green bg-green-lightest rounded-sm">` + SafeHTMLForCommandResult(result) + `</div>`)
						}
					} else {
						return htmlError(err)
					}
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
<div class="my-6">
{{displayCommandResult .}}
</div>
{{end}}

{{define "postActions"}}
<div class="mt-4" data-target="posts.actions">
	<form data-target="posts.createReplyForm" method="post" action="{{childPostsURL .Key.Encode}}" class="my-4"></form>
	<div class="flex row justify-between">
		<div></div>
		<div class="flex row bg-grey-lightest border border-grey-lightest rounded">
			<button data-action="posts#beginReply" class="px-2 py-1 text-grey-darkest"> ↩︎</button>
			<button data-action="posts#addToFaves" class="px-2 py-1 text-grey-darkest border-l border-grey-lighter"> ☆</button>
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
<div class="p-4 pb-6 bg-white border-t-2 border-blue-light rounded-sm shadow" data-target="posts.post">
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
<div class="p-4 pb-6 bg-white border-t-4 border-blue-light rounded-sm" data-target="posts.post">
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
<textarea data-target="posts.mainTextarea" data-action="input->posts#markdownInputChanged" name="markdownSource" rows="4" placeholder="Write…" class="block w-full p-2 bg-white border border-grey rounded shadow-inner"></textarea>
<div class="flex flex-row-reverse">
<button type="submit" name="action" value="submitPost" data-target="posts.submitPostButton" class="mt-2 px-4 py-2 font-bold text-white bg-indigo-darker border border-indigo-darker rounded shadow">Post</button>
<button type="submit" name="action" value="runCommand" data-target="posts.runCommandButton" class="mt-2 px-4 py-2 font-bold text-green-dark bg-white border border-green-dark rounded shadow hidden">Run</button>
<button type="submit" name="action" value="beginDraft" data-target="posts.beginDraftButton" class="mt-2 px-4 py-2 font-bold text-white bg-purple-dark border border-purple-dark rounded shadow hidden">Begin Draft</button>
<button type="submit" name="action" value="runGraphQLQuery" data-target="posts.runGraphQLQueryButton" class="mt-2 px-4 py-2 font-bold text-white bg-pink-dark border border-pink-dark rounded shadow hidden">Run GraphQL Query</button>
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
	<summary class="max-w-md mx-auto p-1 italic cursor-pointer text-center sm:text-right text-sm text-indigo-lighter bg-indigo-darker select-none">Developer</summary>
	<div class="sm:max-h-screen overflow-auto bg-yellow-lightest">
		<div class="flex flex-col sm:flex-row">
			<pre class="sm:w-1/2 p-2 bg-indigo-lightest text-indigo-darkest"><code data-target="developer.queryCode">` + query + `</code></pre>
			<div class="sm:w-1/2">
				<button data-action="developer#runQuery" class="w-full mb-1 px-4 py-2 whitespace-pre-wrap break-words font-bold text-white bg-green-darker border border-green-darker">►</button>
				<pre class="p-2 whitespace-pre-wrap break-words text-green-darkest"><code data-target="developer.result"></code></pre>
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

		sw.WriteString(`<div class="mx-2 md:mx-0">`)
		viewCreatePostFormInChannelHTMLHandle(channelViewModel, sw)
		sw.WriteString(`</div>`)

		sw.WriteString(`<div class="mb-6">`)
		viewPostsInChannelHTMLHandle(ctx, posts, channelViewModel, sw)
		sw.WriteString(`</div>`)

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

			sw.WriteString(`<div class="mt-4">`)
			viewPostInChannelHTMLHandle(ctx, *post, channelViewModel, sw)
			sw.WriteString(`</div>`)

			sw.WriteString(`<div hidden class="hidden">`)
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
	} else if action == "runGraphQLQuery" {
		commandType = "v0-runGraphQLQuery"
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
