package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	// "github.com/BurntSushi/toml"
)

// ParseColorCommand parses a /color … command
func ParseColorCommand(subcommands []string, params string) (Command, error) {
	if len(subcommands) == 1 {
		if subcommands[0] == "gradient" {
			return ParseColorGradientCommand(params)
		}
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

// A ColorGradientCommand represents the `/color gradient` command
type ColorGradientCommand struct {
	Inputs []string
	Colors []colorful.Color
}

// ParseColorGradientCommand creates a new `/color gradient` command
func ParseColorGradientCommand(params string) (*ColorGradientCommand, error) {
	lines := strings.Split(params, "\n")

	cmd := ColorGradientCommand{
		Inputs: lines,
	}

	var colors []colorful.Color
	for _, line := range lines {
		color, err := colorful.Hex(line)
		if err != nil {
			return nil, err
		}
		colors = append(colors, color)
	}

	cmd.Colors = colors

	return &cmd, nil
}

// Run converts the color to a preview
func (cmd *ColorGradientCommand) Run(ctx context.Context) (CommandResult, error) {
	var gradientStops []string
	for _, color := range cmd.Colors {
		hex := color.Hex()
		gradientStops = append(gradientStops, hex)
	}

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<div style="width: 12em; height: 12em; background: linear-gradient(` + strings.Join(gradientStops, ",") + `)"></div>`)
	htmlBuffer.WriteString(`<dl class="mt-4">`)
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">Hex</dt><dd>%s</dd>`, strings.Join(gradientStops, ", ")))
	htmlBuffer.WriteString(`</dl>`)

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}
