package main

import (
	"encoding/json"
	// "fmt"
	// "log"
	"net/http"

	// "cloud.google.com/go/datastore"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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
	}

	writeJSON(w, &struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

func main() {
	r := mux.NewRouter()

	r.Path("/").Methods("GET").
		HandlerFunc(rootHandle)

	r.Path("/user-credentials").Methods("POST").
		HandlerFunc(createUserCredentialHandle)

	http.Handle("/", r)

	appengine.Main()
}
