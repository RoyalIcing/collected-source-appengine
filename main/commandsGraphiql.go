package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/BurntSushi/toml"
)

// ParseGraphiqlCommand parses a /graphiql … command
func ParseGraphiqlCommand(subcommands []string, params string) (Command, error) {
	if len(subcommands) == 0 {
		return ParseGraphiqlMainCommand(params)
	}

	return nil, fmt.Errorf("Unknown graphiql subcommand(s) %v", subcommands)
}

// A GraphiqlMainCommand represents the `/graphiql` command
type GraphiqlMainCommand struct {
	EndpointURL string            `toml:"endpoint"`
	Headers     map[string]string `toml:"headers"`
}

// ParseGraphiqlMainCommand creates a new `/graphiql` command
func ParseGraphiqlMainCommand(params string) (*GraphiqlMainCommand, error) {
	var cmd GraphiqlMainCommand

	_, err := toml.Decode(params, &cmd)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

// Run shows a Graphiql editor for the passed endpoint
func (cmd *GraphiqlMainCommand) Run(ctx context.Context) (CommandResult, error) {
	t := template.New("graphiql command")
	t = template.Must(t.Parse(`
{{define "result"}}
<div id="collected-graphiql-command-result" style="height: 1000px;"></div>
<script>
window.collectedTasks.push({
	method: 'renderGraphiqlForURL',
	params: {
		domElement: document.getElementById('collected-graphiql-command-result'),
		endpointURL: "{{ .EndpointURL }}",
		headers: {{ .Headers }}
	}
})
</script>
{{end}}
	`))

	var htmlBuffer bytes.Buffer
	t.ExecuteTemplate(&htmlBuffer, "result", cmd)
	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	// result.wantsFullWidth = true
	result.SetFullWidth(true)

	return result, nil
}
