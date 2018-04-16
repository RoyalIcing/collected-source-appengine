package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	baseLog "log"
	"net/http"
	"os"
	"time"

	// "cloud.google.com/go/datastore"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/icza/session"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	oauthGitHub "golang.org/x/oauth2/github"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	// "google.golang.org/appengine/memcache"
)

var (
	oauthCfg *oauth2.Config

	gitHubScopes   = []string{"user", "repo"}
	gitHubStateKey = "gitHubState"
)

// UserCredential stores GitHub, Trello API long term keys
type UserCredential struct {
	Email string
}

func writeJSON(w http.ResponseWriter, d interface{}) {
	w.Header().Set("Content-Type", "application/json")

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

func getSessionManager(ctx context.Context) session.Manager {
	return session.NewCookieManagerOptions(session.NewMemcacheStore(ctx), &session.CookieMngrOptions{
		CookieMaxAge: time.Hour * time.Duration(24*30*6), // 6 months
		AllowHTTP:    appengine.IsDevAppServer(),
	})
}

func gitHubOAuthStartHandle(w http.ResponseWriter, r *http.Request) {
	baseLog.Print("gitHubOAuthStartHandle()")
	ctx := appengine.NewContext(r)
	log.Debugf(ctx, "gitHubOAuthStartHandle()")

	stateBytes := make([]byte, 16)
	_, err := rand.Read(stateBytes)
	log.Debugf(ctx, "Generated random state")
	if err != nil {
		log.Errorf(ctx, "Could not generate random state")
		writeErrorJSON(w, err)
		return
	}

	state := base64.URLEncoding.EncodeToString(stateBytes)

	sessmgr := getSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess != nil {
		log.Debugf(ctx, "Set state on existing session")
		sess.SetAttr(gitHubStateKey, state)
	} else {
		sess = session.NewSessionOptions(&session.SessOptions{
			CAttrs: map[string]interface{}{},
			Attrs:  map[string]interface{}{gitHubStateKey: state},
		})
		sessmgr.Add(sess, w)
	}

	url := oauthCfg.AuthCodeURL(state)
	log.Debugf(ctx, "Redirecting to GitHub: %v", url)
	http.Redirect(w, r, url, 302)
}

func githubOAuthCallbackHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sessmgr := getSessionManager(ctx)
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

	sess.SetAttr("gitHubToken", token)

	w.Write([]byte("Success!"))
}

func githubListReposHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sessmgr := getSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess == nil {
		http.Error(w, "You must first sign in.", http.StatusUnauthorized)
		return
	}

	token, ok := sess.Attr("gitHubToken").(oauth2.Token)
	if !ok {
		http.Error(w, "You need to sign in with Github.", http.StatusUnauthorized)
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

func init() {
	gob.Register(oauth2.Token{})
}

func main() {
	if appengine.IsDevAppServer() {
		err := godotenv.Load()
		if err != nil {
			baseLog.Fatal("Error loading .env file")
		}
	}

	oauthCfg = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint:     oauthGitHub.Endpoint,
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		Scopes:       gitHubScopes,
	}

	r := mux.NewRouter()

	r.Path("/").Methods("GET").
		HandlerFunc(rootHandle)

	r.Path("/user-credentials").Methods("POST").
		HandlerFunc(createUserCredentialHandle)

	r.Path("/signin/github").Methods("GET").
		HandlerFunc(gitHubOAuthStartHandle)

	r.Path("/signin/github/callback").Methods("GET").
		HandlerFunc(githubOAuthCallbackHandle)

	r.Path("/github/repos").Methods("GET").
		HandlerFunc(githubListReposHandle)

	r.Path("/_sessions/purge").Methods("GET").
		HandlerFunc(session.PurgeExpiredSessFromDSFunc(""))

	http.Handle("/", r)

	appengine.Main()
}
