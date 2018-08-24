package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RouteVars makes working with URL param vars convenient
type RouteVars struct {
	vars map[string]string
}

func routeVarsFrom(r *http.Request) RouteVars {
	vars := mux.Vars(r)
	return RouteVars{vars}
}

func (v RouteVars) orgSlug() string {
	return v.vars["orgSlug"]
}

func (v RouteVars) channelSlug() string {
	return v.vars["channelSlug"]
}

func (v RouteVars) postID() string {
	return v.vars["postID"]
}

func (v RouteVars) optionalPostID() *string {
	postID, ok := v.vars["postID"]
	if ok {
		return &postID
	} else {
		return nil
	}
}
