package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
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

	figmaAPIBaseURL = "https://api.figma.com"
)

var (
	figmaOauthCfg *oauth2.Config
	figmaScopes   = []string{"file_read"}
)

// FigmaAPI allows retrieving data from the Figma API
type FigmaAPI struct {
	client *http.Client
	token  oauth2.Token
}

func (figma *FigmaAPI) get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", figmaAPIBaseURL+path, nil)
	if err != nil {
		return nil, errors.New("Unable to make request for Figma. " + err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+figma.token.AccessToken)

	return figma.client.Do(req)
}

// ReadFile loads the contents of a particular file from Figma
func (figma *FigmaAPI) ReadFile(key string) (*http.Response, error) {
	return figma.get("/v1/files/" + key)
}

func init() {
	clientID := os.Getenv("FIGMA_CLIENT_ID")
	clientSecret := os.Getenv("FIGMA_CLIENT_SECRET")

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

// GetFigmaAPIFromSession returns a http.Client from a session
func GetFigmaAPIFromSession(ctx context.Context, sess session.Session) *FigmaAPI {
	token, ok := sess.Attr(figmaTokenKey).(oauth2.Token)
	if !ok {
		return nil
	}

	client := urlfetch.Client(ctx)

	return &FigmaAPI{
		client: client,
		token:  token,
	}
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

	figmaAPI := GetFigmaAPIFromSession(ctx, sess)
	if figmaAPI == nil {
		http.Error(w, "You need to sign in with Figma.", http.StatusUnauthorized)
	}

	key := mux.Vars(r)["key"]
	resp, err := figmaAPI.ReadFile(key)
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
