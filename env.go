package main

import (
	baseLog "log"

	"github.com/joho/godotenv"
	"google.golang.org/appengine"
)

var _ = loadEnvIfNeeded()

func loadEnvIfNeeded() (err error) {
	if appengine.IsDevAppServer() {
		err := godotenv.Load()
		if err != nil {
			baseLog.Fatal("Error loading .env file")
		}
		return err
	}

	return nil
}
