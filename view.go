package main

import (
	"bufio"
	"html/template"
	"io"
)

// OrgViewModel models viewing orgs
type OrgViewModel struct {
	OrgSlug string
}

// ViewNav renders the primary navigation
func (m OrgViewModel) ViewNav(w *bufio.Writer) {
	t := template.Must(template.New("nav").Parse(`
<nav class="mb-8 text-white bg-blue-darkest">
<div class="max-w-md mx-auto flex flex-row">
<strong class="py-4">{{.OrgSlug}}</strong>
</div>
</nav>
`))

	t.Execute(w, m)
}

// ViewPage renders a page with navigation and provided main content
func (m OrgViewModel) ViewPage(w io.Writer, viewMainContent func(w *bufio.Writer)) {
	sw := bufio.NewWriter(w)
	defer sw.Flush()

	m.ViewNav(sw)

	sw.WriteString(`<main class="max-w-md mx-auto mt-6 mb-8">`)
	viewMainContent(sw)
	sw.WriteString(`</main>`)
}

// ToOrgViewModel converts route vars into OrgViewModel
func (v RouteVars) ToOrgViewModel() OrgViewModel {
	return OrgViewModel{
		OrgSlug: v.orgSlug(),
	}
}
