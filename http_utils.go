package main

import (
	"context"
	"net/http"
	"time"

	"github.com/icza/session"
	"google.golang.org/appengine"
)

const (
	sessionDuration = time.Hour * time.Duration(24*30*6) // 6 months
)

// GetSessionManager allows storing data within a session
func GetSessionManager(ctx context.Context) session.Manager {
	return session.NewCookieManagerOptions(session.NewMemcacheStore(ctx), &session.CookieMngrOptions{
		CookieMaxAge: sessionDuration,
		AllowHTTP:    appengine.IsDevAppServer(),
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
func WithSessionMgr(f func(
	context.Context, http.ResponseWriter, *http.Request, session.Manager,
)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		sessionmgr := GetSessionManager(ctx)
		defer sessionmgr.Close()
		f(ctx, w, r, sessionmgr)
	})
}
