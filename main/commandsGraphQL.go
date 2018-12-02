package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	// "golang.org/x/net/html"
	"google.golang.org/appengine/urlfetch"

	"github.com/BurntSushi/toml"
)

// ParseGraphQLCommand parses a /graphql … command
func ParseGraphQLCommand(subcommands []string, params string) (Command, error) {
	if len(subcommands) == 0 {
		return ParseGraphQLRemoteQueryCommand(params)
	}

	return nil, fmt.Errorf("Unknown graphql subcommand(s) %v", subcommands)
}

// A GraphQLRemoteQueryCommand represents the `/web snippet` command
type GraphQLRemoteQueryCommand struct {
	EndpointURL string            `toml:"endpoint"`
	Query       string            `toml:"query"`
	Headers     map[string]string `toml:"headers"`
}

// ParseGraphQLRemoteQueryCommand creates a new `/graphql` command
func ParseGraphQLRemoteQueryCommand(params string) (*GraphQLRemoteQueryCommand, error) {
	var cmd GraphQLRemoteQueryCommand

	_, err := toml.Decode(params, &cmd)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

// Run queries the remote GraphQL server
func (cmd *GraphQLRemoteQueryCommand) Run(ctx context.Context) (CommandResult, error) {
	client := urlfetch.Client(ctx)

	bodyJSON := &struct {
		Query string `json:"query"`
	}{
		Query: cmd.Query,
	}

	bodyBytes, err := json.Marshal(bodyJSON)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	request, err := http.NewRequest("POST", cmd.EndpointURL, bodyBuffer)
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	for key, value := range cmd.Headers {
		headers.Set(key, value)
	}
	request.Header = headers

	// Perform the GraphQL query
	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resultJSON interface{}
	resultJSONDecoder := json.NewDecoder(res.Body)
	err = resultJSONDecoder.Decode(&resultJSON)
	if err != nil {
		return nil, err
	}

	resultJSONBytes, err := json.MarshalIndent(resultJSON, "", "  ")
	if err != nil {
		return nil, err
	}

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<pre class="whitespace-pre-wrap break-words">`)
	_, err = htmlBuffer.Write(resultJSONBytes)
	if err != nil {
		return nil, err
	}
	htmlBuffer.WriteString("</pre>")

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}
