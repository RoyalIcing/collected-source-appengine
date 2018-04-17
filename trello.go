package main

import (
	"context"
	"encoding/gob"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/icza/session"
	"github.com/mrjones/oauth"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	// "google.golang.org/appengine/log"
)

const (
	scope                 = "read,write"
	trelloRequestTokenKey = "trelloRequestToken"
	trelloAccessTokenKey  = "trelloAccessToken"
)

func init() {
	gob.Register(oauth.RequestToken{})
	gob.Register(oauth.AccessToken{})
}

func makeConsumer(ctx context.Context) *oauth.Consumer {
	consumer := oauth.NewCustomHttpClientConsumer(
		os.Getenv("TRELLO_API_KEY"),
		os.Getenv("TRELLO_API_SECRET"),
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://trello.com/1/OAuthGetRequestToken",
			AuthorizeTokenUrl: "https://trello.com/1/OAuthAuthorizeToken",
			AccessTokenUrl:    "https://trello.com/1/OAuthGetAccessToken",
		},
		urlfetch.Client(ctx),
	)

	consumer.AdditionalAuthorizationUrlParams["name"] = os.Getenv("TRELLO_APP_NAME")
	consumer.AdditionalAuthorizationUrlParams["expiration"] = "never"
	consumer.AdditionalAuthorizationUrlParams["scope"] = scope

	return consumer
}

func oAuthStartHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	consumer := makeConsumer(ctx)

	requestToken, url, err := consumer.GetRequestTokenAndUrl(os.Getenv("TRELLO_REDIRECT_URL"))
	if err != nil {
		http.Error(w, "Could not sign into Trello "+err.Error(), http.StatusFailedDependency)
		return
	}

	sessmgr := GetSessionManager(ctx)
	defer sessmgr.Close()

	sess := sessmgr.Get(r)
	if sess != nil {
		sess.SetAttr(trelloRequestTokenKey, requestToken)
	} else {
		sess = session.NewSessionOptions(&session.SessOptions{
			CAttrs: map[string]interface{}{},
			Attrs:  map[string]interface{}{trelloRequestTokenKey: requestToken},
		})
		sessmgr.Add(sess, w)
	}

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

	requestToken, ok := sess.Attr(trelloRequestTokenKey).(oauth.RequestToken)
	if !ok {
		http.Error(w, "No request token present for signing into Trello. Please try again.", http.StatusExpectationFailed)
		return
	}

	verificationCode := r.URL.Query().Get("oauth_verifier")
	consumer := makeConsumer(ctx)
	accessToken, err := consumer.AuthorizeToken(&requestToken, verificationCode)
	if err != nil {
		http.Error(w, "Could not get Trello token. Please try again.", http.StatusExpectationFailed)
		return
	}

	sess.SetAttr(trelloAccessTokenKey, accessToken)

	w.Write([]byte("Success!"))
}

func addContext(f func(http.ResponseWriter, *http.Request, context.Context)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		f(w, r, ctx)
	})
}

func useSession(f http.HandlerFunc) http.HandlerFunc {
	return f
}

func readProfileHandle(ctx context.Context, w http.ResponseWriter, r *http.Request, sessmgr session.Manager) {
	sess := sessmgr.Get(r)
	if sess == nil {
		http.Error(w, "You must first sign in.", http.StatusUnauthorized)
		return
	}

	accessToken, ok := sess.Attr("trelloAccessToken").(oauth.AccessToken)
	if !ok {
		http.Error(w, "You need to sign in with Trello.", http.StatusUnauthorized)
		return
	}

	consumer := makeConsumer(ctx)
	client, err := consumer.MakeHttpClient(&accessToken)
	if err != nil {
		http.Error(w, "Unable to communicate with Trello. "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := client.Get("https://trello.com/1/members/me")
	if err != nil {
		http.Error(w, "Unable to communicate with Trello. "+err.Error(), http.StatusFailedDependency)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Could not load user profile from Trello. "+err.Error(), http.StatusFailedDependency)
		return
	}

	w.Write(data)
}

// func listBoardsHandle(w http.ResponseWriter, r *http.Request) {
// 	ctx := appengine.NewContext(r)

// 	sessmgr := GetSessionManager(ctx)
// 	defer sessmgr.Close()

// 	sess := sessmgr.Get(r)
// 	if sess == nil {
// 		http.Error(w, "You must first sign in.", http.StatusUnauthorized)
// 		return
// 	}

// 	token, ok := sess.Attr(trelloAccessTokenKey).(oauth2.Token)
// 	if !ok {
// 		http.Error(w, "You need to sign in with Github.", http.StatusUnauthorized)
// 		return
// 	}

// 	client := github.NewClient(oauthCfg.Client(ctx, &token))

// 	user, _, err := client.Users.Get(ctx, "")
// 	if err != nil {
// 		http.Error(w, "No GitHub user found. "+err.Error(), http.StatusExpectationFailed)
// 	}

// 	opt := &github.RepositoryListOptions{Type: "all", Sort: "full_name"}
// 	repos, _, err := client.Repositories.List(ctx, user.GetLogin(), opt)
// 	if err != nil {
// 		writeErrorJSON(w, err)
// 		return
// 	}

// 	writeJSON(w, repos)
// }

// AddTrelloRoutes adds routes for signing in and reading from GitHub
func AddTrelloRoutes(r *mux.Router) {
	r.Path("/signin/trello").Methods("GET").
		HandlerFunc(oAuthStartHandle)

	r.Path("/signin/trello/callback").Methods("GET").
		HandlerFunc(oAuthCallbackHandle)

	r.Path("/trello/profile").Methods("GET").
		HandlerFunc(WithSessionMgr(readProfileHandle))
}
