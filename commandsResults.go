package main

import (
	"github.com/microcosm-cc/bluemonday"
)

// A CommandResult is the result of a command
type CommandResult interface {
	PlainText() string
	UnsafeHTML() string
}

// SafeHTMLForCommandResult santizes HTML
func SafeHTMLForCommandResult(r CommandResult) string {
	p := bluemonday.UGCPolicy()
	raw := r.UnsafeHTML()
	return p.Sanitize(raw)
}

// A HTMLCommandResult holds HTML
type HTMLCommandResult struct {
	unsafeHTML string
}

// UnsafeHTML returns the underlying html.Node
func (r HTMLCommandResult) UnsafeHTML() string {
	return r.unsafeHTML
}

// PlainText returns the result as an unformatted string
func (r HTMLCommandResult) PlainText() string {
	return ""
}
