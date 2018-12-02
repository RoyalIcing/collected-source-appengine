package main

import (
	"github.com/microcosm-cc/bluemonday"
)

// CommandResultHTML holds potential unsafe HTML
type CommandResultHTML struct {
	unsafeHTML     string
	isActuallySafe bool
}

// A CommandResult is the result of a command
type CommandResult interface {
	WantsFullWidth() bool
	PlainText() string
	HTML() *CommandResultHTML
}

// SafeHTMLForCommandResult santizes HTML
func SafeHTMLForCommandResult(r CommandResult) string {
	html := r.HTML()
	var safe string
	if html.isActuallySafe {
		safe = html.unsafeHTML
	} else {
		p := bluemonday.UGCPolicy()
		safe = p.Sanitize(html.unsafeHTML)
	}
	return safe
}

// HTMLCommandResult is a standard HTML result
type HTMLCommandResult struct {
	html           *CommandResultHTML
	wantsFullWidth bool
}

// HTMLCommandResultFrom makes a result fronm unsafe HTML
func HTMLCommandResultFrom(unsafeHTML string) *HTMLCommandResult {
	html := &CommandResultHTML{unsafeHTML: unsafeHTML, isActuallySafe: false}
	return &HTMLCommandResult{html: html}
}

// DangerousHTMLCommandResultFromSafe makes a result fronm safe HTML
func DangerousHTMLCommandResultFromSafe(safeHTML string) *HTMLCommandResult {
	html := &CommandResultHTML{unsafeHTML: safeHTML, isActuallySafe: true}
	return &HTMLCommandResult{html: html}
}

// SetFullWidth to set true for full width, false for slim width
func (r *HTMLCommandResult) SetFullWidth(flag bool) {
	r.wantsFullWidth = flag
}

// WantsFullWidth returns true for full width, false for slim width
func (r *HTMLCommandResult) WantsFullWidth() bool {
	return r.wantsFullWidth
}

// PlainText returns the result as an unformatted string
func (r *HTMLCommandResult) PlainText() string {
	return ""
}

// HTML returns the result as HTML
func (r *HTMLCommandResult) HTML() *CommandResultHTML {
	return r.html
}
