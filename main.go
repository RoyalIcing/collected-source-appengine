package main

import (
	"context"
	"encoding/gob"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/icza/session"
	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	// "google.golang.org/appengine/memcache"
)

// UserCredential stores GitHub, Trello API long term keys
type UserCredential struct {
	Email string
}

func rootHandle(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, &struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

func createUserCredential(ctx context.Context) error {
	kind := "UserCredential"
	key := datastore.NewIncompleteKey(ctx, kind, nil)

	email := "text@example.com"

	userCredential := UserCredential{
		Email: email,
	}

	key, err := datastore.Put(ctx, key, &userCredential)
	if err != nil {
		log.Errorf(ctx, "Failed to save user: %v", err)
		return err
	}

	log.Debugf(ctx, "Successfully put user credential: %v", key)

	return nil
}

func createUserCredentialHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	err := createUserCredential(ctx)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, &struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

func afterSignInHandle(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, os.Getenv("POST_SIGN_IN_URL"), 302)
}

func init() {
	gob.Register(oauth2.Token{})
}

func main() {
	r := mux.NewRouter()

	// r.Path("/").Methods("GET").
	// 	HandlerFunc(rootHandle)

	r.Path("/user-credentials").Methods("POST").
		HandlerFunc(createUserCredentialHandle)

	AddGitHubRoutes(r)
	AddTrelloRoutes(r)
	AddFigmaRoutes(r)

	AddHTMLDashboardRoutes(r)
	AddHTMLPostsRoutes(r)

	AddOrgsRoutes(r)
	AddPostsRoutes(r)

	http.HandleFunc("/auth/status", AuthStatusHandle)

	resolver := NewDataStoreResolver()
	schema := MakeSchema(&resolver)

	graphqlHandler := relay.Handler{Schema: schema}
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		r = r.WithContext(ctx)
		graphqlHandler.ServeHTTP(w, r)
	})

	r.Path("/_sessions/purge").Methods("GET").
		HandlerFunc(session.PurgeExpiredSessFromDSFunc(""))

	http.Handle("/", r)

	appengine.Main()
}
