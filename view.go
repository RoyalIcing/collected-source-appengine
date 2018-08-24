package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
)

// OrgViewModel models viewing an org
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
