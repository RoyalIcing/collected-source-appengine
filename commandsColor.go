package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/lucasb-eyer/go-colorful"
	// "github.com/BurntSushi/toml"
)

// ParseColorCommand parses a /color … command
func ParseColorCommand(subcommands []string, params string) (Command, error) {
	if len(subcommands) == 1 {
		return ParseColorHexCommand(subcommands[0])
	}

	return nil, fmt.Errorf("Unknown subcommand(s) %v", subcommands)
}

// A ColorCommand represents the `/color` command
type ColorCommand struct {
	Input string
	Color colorful.Color
}

// ParseColorHexCommand creates a new `/color {input}` command
func ParseColorHexCommand(input string) (*ColorCommand, error) {
	cmd := ColorCommand{
		Input: input,
	}

	color, err := colorful.Hex(input)
	if err != nil {
		return nil, err
	}

	cmd.Color = color

	return &cmd, nil
}

// Run converts the color to a preview
func (cmd *ColorCommand) Run(ctx context.Context) (CommandResult, error) {
	hex := cmd.Color.Hex()

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<div style="width: 12em; height: 12em; background-color:` + hex + `"></div>`)
	htmlBuffer.WriteString(`<dl class="mt-4">`)
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">Hex</dt><dd>%s</dd>`, hex))
	red, green, blue := cmd.Color.RGB255()
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">sRGB</dt><dd>rgb(%v, %v, %v)</dd>`, red, green, blue))
	l, a, b := cmd.Color.Lab()
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">Lab</dt><dd>lab(%v %v %v)</dd>`, l, a, b))
	htmlBuffer.WriteString(`</dl>`)

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}
