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
	if len(subcommands) == 1 {
		if subcommands[0] == "snippet" {
			return ParseWebSnippetCommand(params)
		} else if subcommands[0] == "meta" {
			return ParseWebMetaCommand(params)
		}
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

	result := HTMLCommandResultFromUnsafe(htmlBuffer.String())

	return result, nil
}

// A WebMetaCommand represents the `/web meta` command
type WebMetaCommand struct {
	URL string `toml:"url"`
}

// ParseWebMetaCommand creates a new `/web meta` command
func ParseWebMetaCommand(params string) (*WebMetaCommand, error) {
	var cmd WebMetaCommand

	_, err := toml.Decode(params, &cmd)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

// Run fetches the web page and extracts the meta tags from it
func (cmd *WebMetaCommand) Run(ctx context.Context) (CommandResult, error) {
	client := urlfetch.Client(ctx)
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
	metaTags := dom.Find("head meta")

	htmlNodes := metaTags.Nodes

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<ol>`)
	for _, node := range htmlNodes {
		htmlBuffer.WriteString(`<li class="mb-2">`)
		htmlBuffer.WriteString(`<dl class="grid-1/3-2/3 grid-column-gap-1 grid-row-gap-1">`)
		for _, attr := range node.Attr {
			htmlBuffer.WriteString(fmt.Sprintf(`<dt class="font-bold">%s</dt>`, html.EscapeString(attr.Key)))
			htmlBuffer.WriteString(fmt.Sprintf(`<dd>%s</dd>`, html.EscapeString(attr.Val)))
		}
		htmlBuffer.WriteString(`</dl>`)
		htmlBuffer.WriteString(`</li>`)
	}
	htmlBuffer.WriteString(`</ol>`)

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}
