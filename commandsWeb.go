package main

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"golang.org/x/net/html"
	"google.golang.org/appengine/urlfetch"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
)

// ParseWebCommand parses a /web … command
func ParseWebCommand(subcommands []string, params string) (Command, error) {
	if (len(subcommands) == 1) && (subcommands[0] == "snippet") || true {
		return ParseWebSnippetCommand(params)
	}

	return nil, fmt.Errorf("Unknown subcommand(s) %v", subcommands)
}

// A WebSnippetCommand represents the `/web snippet` command
type WebSnippetCommand struct {
	URL      string  `toml:"url"`
	Selector *string `toml:"selector"`
}

// ParseWebSnippetCommand creates a new `/web snippet` command
func ParseWebSnippetCommand(params string) (*WebSnippetCommand, error) {
	var cmd WebSnippetCommand

	_, err := toml.Decode(params, &cmd)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

// Run fetches the web page and extracts a snippet of it
func (cmd *WebSnippetCommand) Run(ctx context.Context) (CommandResult, error) {
	client := urlfetch.Client(ctx)
	url, err := url.Parse(cmd.URL)
	if err != nil {
		return nil, err
	}
	// Request the HTML page.
	res, err := client.Get(cmd.URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	dom := doc.Clone()
	if cmd.Selector != nil {
		dom = doc.Find(*cmd.Selector)
	}

	dom.Parent().Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			linkURL, _ := url.Parse(href)
			if linkURL != nil {
				linkURL = url.ResolveReference(linkURL)
				s.SetAttr("href", linkURL.String())
			}
		}
	})

	htmlNodes := dom.Nodes

	var htmlBuffer bytes.Buffer
	for _, node := range htmlNodes {
		html.Render(&htmlBuffer, node)
		htmlBuffer.WriteString("<br>")
	}

	result := HTMLCommandResult{htmlBuffer.String()}

	return result, nil
}
