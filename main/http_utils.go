package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/icza/session"
	"google.golang.org/appengine"
)

const (
	sessionDuration = time.Hour * time.Duration(24*30*6) // 6 months
)

// SessHandlerFunc has context.Context, session.Manager as extra arguments to a http.HandlerFunc
type SessHandlerFunc func(context.Context, http.ResponseWriter, *http.Request, session.Manager)

// GetSessionManager allows storing data within a session
func GetSessionManager(ctx context.Context) session.Manager {
	return session.NewCookieManagerOptions(session.NewMemcacheStoreOptions(ctx, &session.MemcacheStoreOptions{
		OnlyMemcache:       false,
		AsyncDatastoreSave: false,
	}), &session.CookieMngrOptions{
		CookieMaxAge: sessionDuration,
		AllowHTTP:    IsDev(),
	})
}

// WithContext adds context.Context as an extra argument to a http.HandlerFunc
func WithContext(f func(
	context.Context, http.ResponseWriter, *http.Request,
)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		f(ctx, w, r)
	})
}

// WithSessionMgr adds context.Context and session.Manager as extra arguments to a http.HandlerFunc
func WithSessionMgr(f SessHandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		sessmgr := GetSessionManager(ctx)
		defer sessmgr.Close()
		f(ctx, w, r, sessmgr)
	})
}

// WithViewer adds context.Context and Viewer as extra arguments to a http.HandlerFunc
func WithViewer(f func(context.Context, *Viewer, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		sessmgr := GetSessionManager(ctx)
		defer sessmgr.Close()

		sess := sessmgr.Get(r)
		viewer := NewViewer(ctx, sess)

		f(ctx, viewer, w, r)
	})
}

// WithViewerInSession ensures there is a session started, and adds context.Context and Viewer as extra arguments to a http.HandlerFunc
func WithViewerInSession(f func(context.Context, *Viewer, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		sessmgr := GetSessionManager(ctx)
		defer sessmgr.Close()

		sess := sessmgr.Get(r)
		if sess == nil {
			sess = session.NewSessionOptions(&session.SessOptions{
				CAttrs: map[string]interface{}{},
				Attrs:  map[string]interface{}{},
			})
			sessmgr.Add(sess, w)
		}

		viewer := NewViewer(ctx, sess)

		f(ctx, viewer, w, r)
	})
}

func writeJSON(w http.ResponseWriter, d interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// TODO: use json.NewEncoder(w).Encode(...)
	b, err := json.Marshal(d)

	if err != nil {
		http.Error(w, "{\"error\": \"Could not encode json\"}", http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func writeErrorJSON(w http.ResponseWriter, e error) {
	writeJSON(w, &struct {
		Error string `json:"error"`
	}{
		Error: e.Error(),
	})
}
