package main

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

// AddAPIStorageRoutes adds routes for working with storage
func AddAPIStorageRoutes(r *mux.Router) {
	// Text
	r.Path("/1/storage/text/markdown/sha256/{sha256}").Methods("POST").
		HandlerFunc(createTextMarkdownInStorageHandle)
	r.Path("/1/storage/text/markdown/sha256/{sha256}").Methods("GET").
		HandlerFunc(readTextMarkdownInStorageHandle)
	// Images
	r.Path("/1/storage/image/png/sha256/{sha256}").Methods("POST").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			createImageInStorageHandle("image/png", w, r)
		})
	r.Path("/1/storage/image/png/sha256/{sha256}").Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readImageInStorageHandle("image/png", w, r)
		})
	r.Path("/1/storage/image/jpeg/sha256/{sha256}").Methods("POST").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			createImageInStorageHandle("image/jpeg", w, r)
		})
	r.Path("/1/storage/image/jpeg/sha256/{sha256}").Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readImageInStorageHandle("image/jpeg", w, r)
		})
}

func createTextMarkdownInStorageHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	storageRepo := NewStorageRepo(ctx)

	err := storageRepo.addContentWithMediaTypeAndSHA256("text/markdown", vars.sha256(), r.Body)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, "success")
}

func readTextMarkdownInStorageHandle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	storageRepo := NewStorageRepo(ctx)

	reader, err := storageRepo.readContentWithMediaTypeAndSHA256("text/markdown", vars.sha256())
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
	reader.Close()
}

func createImageInStorageHandle(mediaType string, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	storageRepo := NewStorageRepo(ctx)

	err := storageRepo.addContentWithMediaTypeAndSHA256(mediaType, vars.sha256(), r.Body)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	writeJSON(w, "success")
}

func readImageInStorageHandle(mediaType string, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := routeVarsFrom(r)

	storageRepo := NewStorageRepo(ctx)

	reader, err := storageRepo.readContentWithMediaTypeAndSHA256(mediaType, vars.sha256())
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	w.Header().Set("Content-Type", mediaType)
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
	reader.Close()
}
