package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	baseLog "log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/icza/session"
	"github.com/joho/godotenv"
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

// GetSessionManager allows storing data within a session
func GetSessionManager(ctx context.Context) session.Manager {
	return session.NewCookieManagerOptions(session.NewMemcacheStore(ctx), &session.CookieMngrOptions{
		CookieMaxAge: time.Hour * time.Duration(24*30*6), // 6 months
		AllowHTTP:    appengine.IsDevAppServer(),
	})
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

	r := mux.NewRouter()

	r.Path("/").Methods("GET").
		HandlerFunc(rootHandle)

	r.Path("/user-credentials").Methods("POST").
		HandlerFunc(createUserCredentialHandle)

	AddGitHubRoutes(r)

	r.Path("/_sessions/purge").Methods("GET").
		HandlerFunc(session.PurgeExpiredSessFromDSFunc(""))

	http.Handle("/", r)

	appengine.Main()
}
