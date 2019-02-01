package main

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	// "google.golang.org/appengine"
)

// StorageRepo lets you upload content
type StorageRepo struct {
	ctx context.Context
}

// NewStorageRepo makes a new storage repository
func NewStorageRepo(ctx context.Context) StorageRepo {
	return StorageRepo{
		ctx: ctx,
	}
}

func objectForStorageContent(ctx context.Context, key string) (*storage.ObjectHandle, error) {
	bucketName := "staging-syrup-storage"

	// if appengine.IsDevAppServer()
	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadWrite)

	storageClient, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("Cannot make storage client")
	}

	bucket := storageClient.Bucket(bucketName)
	object := bucket.Object(key)

	return object, nil
}

func (repo *StorageRepo) addContentWithMediaTypeAndSHA256(mediaType string, sha256 string, r io.ReadCloser) error {
	key := `mediaType/` + mediaType + `/sha256/` + sha256
	object, err := objectForStorageContent(repo.ctx, key)
	if object == nil {
		return err
	}

	writer := object.NewWriter(repo.ctx)
	writer.ContentType = mediaType

	_, err = io.Copy(writer, r)
	if err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func (repo *StorageRepo) readContentWithMediaTypeAndSHA256(mediaType string, sha256 string) (io.ReadCloser, error) {
	key := `mediaType/` + mediaType + `/sha256/` + sha256
	object, err := objectForStorageContent(repo.ctx, key)
	if object == nil {
		return nil, err
	}

	return object.NewReader(repo.ctx)
}
