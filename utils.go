package main

import (
	"google.golang.org/appengine"
)

// IsDev Returns true if in development, otherwise we are deployed somewhere real
func IsDev() bool {
	return appengine.IsDevAppServer()
}
