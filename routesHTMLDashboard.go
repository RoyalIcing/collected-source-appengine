package main

import (
	"context"
	"html/template"
	"net/http"

	// "html/template"

	"github.com/gorilla/mux"
)

// AddHTMLDashboardRoutes adds routes for the dashboard
func AddHTMLDashboardRoutes(r *mux.Router) {
	r.Path("/").Methods("GET").
		HandlerFunc(WithHTMLTemplate(WithViewer(showDashboardHTMLHandle), htmlHandlerOptions{}))
}

func showDashboardHTMLHandle(ctx context.Context, v *Viewer, w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("dashboardPage").Parse(`
<header class="mt-8 mb-8">
<h1 class="font-lg text-center">Collected</h1>
</h1>
</header>

<main class="mb-8">
	<div class="max-w-md mx-auto">

		<ul class="text-lg leading-normal">
			<li>Collaborate within channels.</li>
			<li>Write short posts, shared privately within your channel.</li>
			<li>Publish longform pieces, shared with your organization or even publically.</li>
			<li>Use Markdown, images, design elements.</li>
		</ul>

		<section class="mt-8">
		{{if .GetGitHubClient}}
		<article class="px-4 py-3 bg-white border border-grey-lighter rounded">
		<p class="text-lg">Signed into GitHub</p>
		</article>
		{{else}}
		<a href="/signin/github" class="px-4 py-2 text-lg tracking-wide text-white bg-black rounded no-underline">Sign into GitHub</a>
		{{end}}
		</section>

	</div>
</main>
`))

	t.Execute(w, v)

	// orgRepo := NewOrgRepo(ctx, vars.orgSlug())
}
