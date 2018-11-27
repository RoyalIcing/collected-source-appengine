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

// SetAlert stores an error message to show the user
func (v *Viewer) SetAlert(errorMessage string) {
	if v.sess == nil {
		return
	}

	v.sess.SetAttr("alert", errorMessage)
}

// ReadAlert reads the previously set error message, if one exists
func (v *Viewer) ReadAlert() *string {
	if v.sess == nil {
		return nil
	}

	errorMessage, ok := v.sess.Attr("alert").(string)
	v.sess.SetAttr("alert", nil)
	if ok {
		return &errorMessage
	}

	return nil
}

// GetGitHubClient returns the github.Client for the signed in user, if there is one
func (v *Viewer) GetGitHubClient() *github.Client {
	if v.sess == nil {
		return nil
	}

	return GetGitHubClientFromSession(v.ctx, v.sess)
}
