package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

func viewErrorMessage(errorMessage string, w *bufio.Writer) {
	w.WriteString(`<p class="py-1 px-2 bg-white text-red">` + errorMessage + "</p>")
}

type htmlHandlerOptions struct {
	form                   bool
	dynamicElementsEnabled map[string]bool
}

func writeDynamicElementsScript(w http.ResponseWriter, dynamicElementsEnabled map[string]bool) {
	if len(dynamicElementsEnabled) == 0 {
		return
	}

	t := template.Must(template.New("dynamicElementsScript").Parse(`
<script>
document.addEventListener("DOMContentLoaded", () => {
const app = Stimulus.Application.start();

{{if .posts}}
app.register('posts', class extends Stimulus.Controller {
	static get targets() {
		return [ 'post', 'replyHolder' ];
	}

	beginReply({ target: button }) {
		const actions = button.closest('[data-target="posts.actions"]');
		const createReplyForm = actions.querySelector('[data-target="posts.createReplyForm"]');
		const createForm = this.targets.find('createForm'); // this.createFormTarget;
		createReplyForm.innerHTML = createForm.innerHTML;
	}
	
	markdownInputChanged({ target: { value } }) {
		const isCommand = value[0] === '/';
		const isMarkdownHeading = value[0] === '#' && value[1] === ' ';
		const isGraphQLQuery = /^query\s+.*{/.test(value);
		this.changeSubmitMode(isCommand ? 'run' : isMarkdownHeading ? 'draft' : isGraphQLQuery ? 'graphQLQuery' : 'submit');
	}

	changeSubmitMode(mode) {
		this.targets.find('submitPostButton').classList.toggle('hidden', mode !== 'submit');
		this.targets.find('runCommandButton').classList.toggle('hidden', mode !== 'run');
		this.targets.find('beginDraftButton').classList.toggle('hidden', mode !== 'draft');
		this.targets.find('runGraphQLQueryButton').classList.toggle('hidden', mode !== 'graphQLQuery');
		this.targets.find('mainTextarea').classList.toggle('font-mono', mode === 'run' || mode === 'graphQLQuery');
	}
});
{{end}}
{{if .developer}}
app.register('developer', class extends Stimulus.Controller {
	static get targets() {
		return [ 'queryCode' ];
	}

	runQuery({ target: button }) {
		const queryCodeEl = this.targets.find('queryCode'); // this.queryCodeTarget;
		const resultEl = this.targets.find('result');
		resultEl.textContent = "Loading…";
		fetch('/graphql', {
			method: 'POST',
			body: JSON.stringify({
				query: queryCodeEl.textContent
			})
		})
			.then(res => res.json())
			.then(json => {
				resultEl.textContent = JSON.stringify(json, null, 2);
			});
	}
});
{{end}}

});
</script>
`))

	t.Execute(w, dynamicElementsEnabled)
}

func WithHTMLTemplate(f http.HandlerFunc, options htmlHandlerOptions) http.HandlerFunc {
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
<meta name="viewport" content="width=device-width, initial-scale=1">
<link href="https://cdn.jsdelivr.net/npm/tailwindcss/dist/tailwind.min.css" rel="stylesheet">
<script defer src="https://unpkg.com/stimulus@1.0.1/dist/stimulus.umd.js"></script>
<style>
.grid-1\/3-2\/3 {
	display: grid;
	grid-template-columns: 33.333% 66.667%;
}
.grid-column-gap-1 {
	grid-column-gap: 0.25rem;
}
.grid-row-gap-1 {
	grid-row-gap: 0.25rem;
}
</style>
</head>
<body class="bg-grey-lightest">
`)

		if formErr != nil {
			w.WriteHeader(400)
			io.WriteString(w, "Invalid form request: "+formErr.Error())
		} else {
			f(w, r)
		}

		writeDynamicElementsScript(w, options.dynamicElementsEnabled)

		io.WriteString(w, "</body></html>")
	})
}

// OrgViewModel models viewing an org
type OrgViewModel struct {
	OrgSlug string
}

func (m OrgViewModel) viewNav(w *bufio.Writer) {
	t := template.Must(template.New("nav").Parse(`
<nav class="text-white bg-black">
<div class="max-w-md mx-auto flex flex-row leading-normal">
<strong class="py-1">{{.OrgSlug}}</strong>
</div>
</nav>
`))

	t.Execute(w, m)
}

// ViewPage renders a page with navigation and provided main content
func (m OrgViewModel) ViewPage(w io.Writer, viewMainContent func(section func(wide bool, viewInner func(w *bufio.Writer)))) {
	sw := bufio.NewWriter(w)
	defer sw.Flush()

	m.viewNav(sw)

	section := func(wide bool, viewInner func(w *bufio.Writer)) {
		if wide {
			sw.WriteString(`<div class="">`)
		} else {
			sw.WriteString(`<div class="max-w-md mx-auto">`)
		}
		viewInner(sw)
		sw.WriteString(`</div>`)
	}

	sw.WriteString(`<main>`)
	viewMainContent(section)
	sw.WriteString(`</main>`)
}

// ChannelViewModel models viewing a channel within an org
type ChannelViewModel struct {
	Org         OrgViewModel
	ChannelSlug string
}

// Channel makes a model for
func (m OrgViewModel) Channel(channelSlug string) ChannelViewModel {
	return ChannelViewModel{
		Org:         m,
		ChannelSlug: channelSlug,
	}
}

// HTMLPostsURL builds a URL to a channel’s posts web page
func (m ChannelViewModel) HTMLPostsURL() string {
	return fmt.Sprintf("/org:%s/channel:%s/posts", m.Org.OrgSlug, m.ChannelSlug)
}

// HTMLPostURL builds a URL to a post
func (m ChannelViewModel) HTMLPostURL(postID string) string {
	return fmt.Sprintf("/org:%s/channel:%s/posts/%s", m.Org.OrgSlug, m.ChannelSlug, postID)
}

// HTMLPostChildPostsURL builds a URL to a post’s child posts web page
func (m ChannelViewModel) HTMLPostChildPostsURL(postID string) string {
	return fmt.Sprintf("/org:%s/channel:%s/posts/%s/posts", m.Org.OrgSlug, m.ChannelSlug, postID)
}

// ToOrgViewModel converts route vars into OrgViewModel
func (v RouteVars) ToOrgViewModel() OrgViewModel {
	return OrgViewModel{
		OrgSlug: v.orgSlug(),
	}
}

// ToChannelViewModel converts route vars into ChannelViewModel
func (v RouteVars) ToChannelViewModel() ChannelViewModel {
	return v.ToOrgViewModel().Channel(v.channelSlug())
}
