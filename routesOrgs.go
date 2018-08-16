package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

func AddOrgsRoutes(r *mux.Router) {
	r.Path("/1/org:{orgName}").Methods("PUT").
		HandlerFunc(createOrgHandle)
}

func createOrgHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	orgRepo := NewOrgRepo(ctx, "RoyalIcing")
	org, err := orgRepo.CreateOrg()
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, org)
}
