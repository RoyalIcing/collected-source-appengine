package main

import (
	"context"
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
	githubOauthCfg *oauth2.Config
	githubScopes   = []string{"user", "repo"}
)

func init() {
	githubOauthCfg = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint:     oauthGitHub.Endpoint,
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		Scopes:       githubScopes,
	}
}

func githubOAuthStartHandle(w http.ResponseWriter, r *http.Request) {
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

	url := githubOauthCfg.AuthCodeURL(state)
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
	token, err := githubOauthCfg.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Could not get GitHub token. Please try again.", http.StatusBadRequest)
		return
	}

	if !token.Valid() {
		http.Error(w, "GitHub token is invalid. Please try again.", http.StatusBadRequest)
		return
	}

	// client := github.NewClient(githubOauthCfg.Client(ctx, token))

	sess.SetAttr(gitHubTokenKey, token)

	w.Write([]byte("Success!"))
}

// GetGitHubClientFromSession returns a github.Client from a session
func GetGitHubClientFromSession(ctx context.Context, sess session.Session) *github.Client {
	token, ok := sess.Attr(gitHubTokenKey).(oauth2.Token)
	if !ok {
		return nil
	}

	return github.NewClient(githubOauthCfg.Client(ctx, &token))
}

// WithGitHubClient adds github.Client as extra arguments to a SessHandlerFunc
func WithGitHubClient(f func(
	context.Context, http.ResponseWriter, *http.Request, *github.Client, session.Manager,
)) SessHandlerFunc {
	return SessHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request, sessmgr session.Manager) {
		sess := sessmgr.Get(r)
		if sess == nil {
			http.Error(w, "You must first sign in.", http.StatusUnauthorized)
			return
		}

		client := GetGitHubClientFromSession(ctx, sess)

		f(ctx, w, r, client, sessmgr)
	})
}

func githubListReposHandle(ctx context.Context, w http.ResponseWriter, r *http.Request, client *github.Client, sessmgr session.Manager) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		http.Error(w, "No GitHub user found. "+err.Error(), http.StatusExpectationFailed)
		return
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
		HandlerFunc(githubOAuthStartHandle)

	r.Path("/signin/github/callback").Methods("GET").
		HandlerFunc(githubOAuthCallbackHandle)

	r.Path("/github/repos").Methods("GET").
		HandlerFunc(WithSessionMgr(WithGitHubClient(githubListReposHandle)))
}
