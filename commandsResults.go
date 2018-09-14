package main

import (
	"github.com/microcosm-cc/bluemonday"
)

// A CommandResult is the result of a command
type CommandResult interface {
	PlainText() string
	UnsafeHTML() string
	DangerousHTMLIsSafe() bool
}

// SafeHTMLForCommandResult santizes HTML
func SafeHTMLForCommandResult(r CommandResult) string {
	p := bluemonday.UGCPolicy()
	raw := r.UnsafeHTML()
	var safe string
	if r.DangerousHTMLIsSafe() {
		safe = raw
	} else {
		safe = p.Sanitize(raw)
	}
	return safe
}

// A HTMLCommandResult holds HTML
type HTMLCommandResult struct {
	unsafeHTML     string
	isActuallySafe bool
}

// HTMLCommandResult makes a result fronm unsafe HTML
func HTMLCommandResultFromUnsafe(unsafeHTML string) HTMLCommandResult {
	return HTMLCommandResult{unsafeHTML, false}
}

// DangerousHTMLCommandResultFromSafe makes a result fronm safe HTML
func DangerousHTMLCommandResultFromSafe(safeHTML string) HTMLCommandResult {
	return HTMLCommandResult{safeHTML, true}
}

// UnsafeHTML returns the underlying html.Node
func (r HTMLCommandResult) UnsafeHTML() string {
	return r.unsafeHTML
}

// DangerousHTMLIsSafe returns whether the underlying HTML is safe
func (r HTMLCommandResult) DangerousHTMLIsSafe() bool {
	return r.isActuallySafe
}

// PlainText returns the result as an unformatted string
func (r HTMLCommandResult) PlainText() string {
	return ""
}
