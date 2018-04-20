package main

import (
	"context"
	"net/http"

	"github.com/icza/session"
	// "google.golang.org/appengine"
)

type authStatusData struct {
	Session bool `json:"session"`
	GitHub  bool `json:"github"`
	Trello  bool `json:"trello"`
	Figma   bool `json:"figma"`
}

// AuthStatusHandle returns json for which services are signed in
func AuthStatusHandle(w http.ResponseWriter, r *http.Request) {
	WithSessionMgr(func(ctx context.Context, w http.ResponseWriter, r *http.Request, sessmgr session.Manager) {
		data := &authStatusData{}
		defer writeJSON(w, data)

		sess := sessmgr.Get(r)
		if sess == nil {
			data.Session = true
			return
		}

		data.Session = true

		gitHubClient := GetGitHubClientFromSession(ctx, sess)
		data.GitHub = gitHubClient != nil

		trelloClient := GetTrelloClientFromSession(ctx, sess)
		data.Trello = trelloClient != nil

		figmaAPI := GetFigmaAPIFromSession(ctx, sess)
		data.Figma = figmaAPI != nil
	})(w, r)
}
