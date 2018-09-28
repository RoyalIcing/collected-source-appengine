package main

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/icza/session"
)

// Viewer represents the current signed in user
type Viewer struct {
	ctx  context.Context
	sess session.Session
}

// NewViewer takes a session and allows getting authenticated services
func NewViewer(ctx context.Context, sess session.Session) *Viewer {
	v := Viewer{ctx, sess}
	return &v
}

// GetGitHubClient returns the github.Client for the signed in user, if there is one
func (v *Viewer) GetGitHubClient() *github.Client {
	if v.sess == nil {
		return nil
	}

	return GetGitHubClientFromSession(v.ctx, v.sess)
}
