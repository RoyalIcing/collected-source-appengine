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
	<div class="max-w-md mx-auto flex flex-row justify-between">
		<div class="text-2xl font-bold">Collected</div>
		<div>
			{{if .GetGitHubClient}}
			<article class="px-4 py-3 bg-white border border-grey-lighter rounded">
			<p class="text-lg">Signed into GitHub</p>
			</article>
			{{else}}
			<a href="/signin/github" class="mt-2 px-4 py-2 font-bold text-white bg-purple-dark border border-purple-darker rounded shadow no-underline hover:bg-purple hover:border-purple-dark">Sign in with GitHub</a>
			{{end}}
		</div>
	</div>
</header>

<main class="mb-8">
	<div class="max-w-md mx-auto">
		<h1 class="mb-4">Share your team’s news with the rest of your organization.</h1>
		<ul class="text-lg leading-normal">
			<li>Collaborate with your organization’s different teams.</li>
			<li>Write privately for your team.</li>
			<li>Curate &amp; refine and then share with your organization.</li>
			<li>Use the same tools to publish to the world.</li>
			<li>Use Markdown, images, GraphQL.</li>
		</ul>
	</div>

	<div class="mt-16 border-b border-purple"></div>

	<div class="max-w-md mx-auto">
		<section class="mt-16">
			<form method="post" action="/org" class="my-4">
			<h2 class="text-purple-dark">Create your team</h2>
			<label class="block my-2">
				<span class="font-bold">Team name</span>
				<input name="orgSlug" class="block w-full mt-1 p-2 bg-grey-lightest border border-grey rounded shadow-inner">
			</label>
			<button type="submit" class="mt-2 px-4 py-2 font-bold text-white bg-purple-dark border border-purple-darker rounded shadow">Create Team</button>
			</form>
		</section>

	</div>
</main>
`))

	t.Execute(w, v)

	// orgRepo := NewOrgRepo(ctx, vars.orgSlug())
}
