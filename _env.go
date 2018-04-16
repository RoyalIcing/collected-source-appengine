package main

import (
	baseLog "log"

	"github.com/joho/godotenv"
	"google.golang.org/appengine"
)

func init() {
	if appengine.IsDevAppServer() {
		err := godotenv.Load()
		if err != nil {
			baseLog.Fatal("Error loading .env file")
		}
	}
}
