package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	// "cloud.google.com/go/datastore"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/icza/session"
	"golang.org/x/oauth2"
	oauthGitHub "golang.org/x/oauth2/github"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const (
	gitHubStateKey = "gitHubState"
	gitHubTokenKey = "gitHubToken"
)

var (
	oauthCfg *oauth2.Config
	scopes   = []string{"user", "repo"}
)

func init() {
	oauthCfg = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint:     oauthGitHub.Endpoint,
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		Scopes:       scopes,
	}
}

func gitHubOAuthStartHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	stateBytes := make([]byte, 16)
	_, err := rand.Read(stateBytes)
	if err != nil {
		log.Errorf(ctx, "Could not generate random state")
		writeErrorJSON(w, err)
		return
	}

	state := base64.URLEncoding.EncodeToString(stateBytes)

	sessmgr := GetSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess != nil {
		sess.SetAttr(gitHubStateKey, state)
	} else {
		sess = session.NewSessionOptions(&session.SessOptions{
			CAttrs: map[string]interface{}{},
			Attrs:  map[string]interface{}{gitHubStateKey: state},
		})
		sessmgr.Add(sess, w)
	}

	url := oauthCfg.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
}

func githubOAuthCallbackHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sessmgr := GetSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess == nil {
		http.Error(w, "No session present for signing in. Please try again.", http.StatusExpectationFailed)
		return
	}

	expectedState := sess.Attr(gitHubStateKey)
	givenState := r.URL.Query().Get("state")
	if expectedState != givenState {
		http.Error(w, "State does not match one at start of sign in flow. Please try again.", http.StatusExpectationFailed)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Could not get GitHub token. Please try again.", http.StatusBadRequest)
		return
	}

	if !token.Valid() {
		http.Error(w, "GitHub token is invalid. Please try again.", http.StatusBadRequest)
		return
	}

	sess.SetAttr(gitHubTokenKey, token)

	w.Write([]byte("Success!"))
}

func githubListReposHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sessmgr := GetSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess == nil {
		http.Error(w, "You must first sign in.", http.StatusUnauthorized)
		return
	}

	token, ok := sess.Attr(gitHubTokenKey).(oauth2.Token)
	if !ok {
		http.Error(w, "You need to sign in with GitHub.", http.StatusUnauthorized)
		return
	}

	client := github.NewClient(oauthCfg.Client(ctx, &token))

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		http.Error(w, "No GitHub user found. "+err.Error(), http.StatusExpectationFailed)
	}

	opt := &github.RepositoryListOptions{Type: "all", Sort: "full_name"}
	repos, _, err := client.Repositories.List(ctx, user.GetLogin(), opt)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, repos)
}

// AddGitHubRoutes adds routes for signing in and reading from GitHub
func AddGitHubRoutes(r *mux.Router) {
	r.Path("/signin/github").Methods("GET").
		HandlerFunc(gitHubOAuthStartHandle)

	r.Path("/signin/github/callback").Methods("GET").
		HandlerFunc(githubOAuthCallbackHandle)

	r.Path("/github/repos").Methods("GET").
		HandlerFunc(githubListReposHandle)
}
