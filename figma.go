package main

import (
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/icza/session"
	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	figmaStateKey = "figmaState"
	figmaTokenKey = "figmaToken"
)

var (
	figmaOauthCfg *oauth2.Config
	figmaScopes   = []string{"file_read"}
)

func init() {
	clientID := os.Getenv("FIGMA_API_KEY")
	clientSecret := os.Getenv("FIGMA_API_SECRET")

	tokenURLQuery := url.Values{}
	tokenURLQuery.Set("client_id", clientID)
	tokenURLQuery.Add("client_secret", clientSecret)
	tokenURL := "https://www.figma.com/api/oauth/token?" + tokenURLQuery.Encode()

	figmaOauthCfg = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.figma.com/oauth",
			TokenURL: tokenURL,
		},
		RedirectURL: os.Getenv("FIGMA_REDIRECT_URL"),
		Scopes:      figmaScopes,
	}
}

func figmaOauthStartHandle(w http.ResponseWriter, r *http.Request) {
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
		sess.SetAttr(figmaStateKey, state)
	} else {
		sess = session.NewSessionOptions(&session.SessOptions{
			CAttrs: map[string]interface{}{},
			Attrs:  map[string]interface{}{figmaStateKey: state},
		})
		sessmgr.Add(sess, w)
	}

	url := figmaOauthCfg.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
}

func oAuthCallbackHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sessmgr := GetSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess == nil {
		http.Error(w, "No session present for signing in. Please try again.", http.StatusExpectationFailed)
		return
	}

	expectedState := sess.Attr(figmaStateKey)
	givenState := r.URL.Query().Get("state")
	if expectedState != givenState {
		http.Error(w, "State does not match one at start of sign in flow. Please try again.", http.StatusExpectationFailed)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := figmaOauthCfg.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Could not get Figma token. Please try again. "+err.Error(), http.StatusBadRequest)
		return
	}

	if !token.Valid() {
		http.Error(w, "Figma token is invalid. Please try again.", http.StatusBadRequest)
		return
	}

	sess.SetAttr(figmaTokenKey, token)

	w.Write([]byte("Success!"))
}

func figmaReadDocumentHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sessmgr := GetSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess == nil {
		http.Error(w, "You must first sign in.", http.StatusUnauthorized)
		return
	}

	token, ok := sess.Attr(figmaTokenKey).(oauth2.Token)
	if !ok {
		http.Error(w, "You need to sign in with Figma.", http.StatusUnauthorized)
		return
	}

	key := mux.Vars(r)["key"]
	client := urlfetch.Client(ctx)

	req, err := http.NewRequest("GET", "https://api.figma.com/v1/files/"+key, nil)
	if err != nil {
		http.Error(w, "Unable to make request for Figma. "+err.Error(), http.StatusInternalServerError)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Unable to load document from Figma. "+err.Error(), resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Could not load document from Figma. "+err.Error(), http.StatusFailedDependency)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// AddFigmaRoutes adds routes for signing in and reading from GitHub
func AddFigmaRoutes(r *mux.Router) {
	r.Path("/signin/figma").Methods("GET").
		HandlerFunc(figmaOauthStartHandle)

	r.Path("/signin/figma/callback").Methods("GET").
		HandlerFunc(oAuthCallbackHandle)

	r.Path("/figma/files/{key}").Methods("GET").
		HandlerFunc(figmaReadDocumentHandle)
}
