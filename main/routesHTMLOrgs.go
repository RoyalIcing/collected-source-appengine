package main

import (
	"bufio"
	"context"
	"net/http"

	// "html/template"

	"github.com/gorilla/mux"
)

// AddHTMLOrgsRoutes adds routes for organizations
func AddHTMLOrgsRoutes(r *mux.Router) {
	r.Path("/org").Methods("POST").
		HandlerFunc(WithHTMLTemplate(WithViewer(createOrgHTMLHandle), htmlHandlerOptions{}))
	r.Path("/org:{orgSlug}").Methods("GET").
		HandlerFunc(WithHTMLTemplate(WithViewer(showOrgHTMLHandle), htmlHandlerOptions{}))
	r.Path("/org:{orgSlug}/channels").Methods("POST").
		HandlerFunc(WithViewerInSession(createChannelInOrgHTMLHandle))
}

func showOrgHTMLHandle(ctx context.Context, v *Viewer, w http.ResponseWriter, r *http.Request) {
	orgViewModel := routeVarsFrom(r).ToOrgViewModel()

	orgRepo := NewOrgRepo(ctx, orgViewModel.OrgSlug)
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	channelsConnections := channelsRepo.NewChannelsConnection(OrgChannelsConnectionOptions{
		maxCount: 100,
	})

	orgViewModel.ViewPage(w, func(viewSection func(wide bool, viewInner func(sw *bufio.Writer))) {
		viewSection(false, func(sw *bufio.Writer) {
			sw.WriteString(`<div class="my-8">`)

			sw.WriteString(`<h2>Channels</h2>`)
			sw.WriteString(`<ul class="list-reset mt-4 rounded shadow">`)
			err := channelsConnections.Enumerate(func(channel ChannelContent) {
				channelViewModel := orgViewModel.Channel(channel.Slug)

				sw.WriteString(`<li class="text-xl">
				<a href="` + channelViewModel.HTMLPostsURL() + `" class="block px-3 py-2 no-underline text-indigo-dark bg-white hover:text-white hover:bg-indigo">#` + channel.Slug + `</a></li>`)
			})
			sw.WriteString(`</ul>`)
			if err != nil {
				viewErrorMessage(err.Error(), sw)
			}

			sw.WriteString(`</div>`)
		})

		viewSection(false, func(sw *bufio.Writer) {
			sw.WriteString(`<div class="my-8">`)

			alert := v.ReadAlert()
			if alert != nil {
				sw.WriteString(`<p class="px-3 py-2 bg-white border-t-4 border-red rounded-sm shadow"><span class="text-red-dark">Error: </span>` + *alert + `</p>`)
			}

			sw.WriteString(`
<form method="post" action="` + orgViewModel.HTMLChannelsURL() + `" class="my-4">
<h2>New Channel</h2>
<label class="block my-2">
	Slug
	<input name="channelSlug" placeholder="e.g. design, engineering, marketing" class="block w-full mt-1 p-2 bg-white border border-grey rounded shadow-inner">
</label>
<button type="submit" class="mt-2 px-4 py-2 font-bold text-white bg-indigo-darker border border-indigo-darker rounded shadow">Create Channel</button>
</form>
`)
			sw.WriteString(`</div>`)
		})
	})
}

func createOrgHTMLHandle(ctx context.Context, v *Viewer, w http.ResponseWriter, r *http.Request) {
	orgSlug := r.PostFormValue("orgSlug")
	orgRepo := NewOrgRepo(ctx, orgSlug)

	org, err := orgRepo.CreateOrg()
	if err != nil {
		v.SetAlert(err.Error())
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	vm := OrgViewModel{OrgSlug: org.Slug}
	http.Redirect(w, r, vm.HTMLURL(), http.StatusFound)
}

func createChannelInOrgHTMLHandle(ctx context.Context, v *Viewer, w http.ResponseWriter, r *http.Request) {
	orgViewModel := routeVarsFrom(r).ToOrgViewModel()

	orgRepo := NewOrgRepo(ctx, orgViewModel.OrgSlug)
	channelsRepo := NewChannelsRepo(ctx, orgRepo)

	channelSlug := r.PostFormValue("channelSlug")

	_, err := channelsRepo.CreateChannel(channelSlug)
	if err != nil {
		v.SetAlert(err.Error())
	}

	http.Redirect(w, r, orgViewModel.HTMLURL(), http.StatusFound)
}
