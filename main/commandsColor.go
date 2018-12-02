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
	input string
	color colorful.Color
}

// Subcommands resolved
func (cmd *ColorCommand) Subcommands() *[]string {
	return nil
}

// Params resolved
func (cmd *ColorCommand) Params() *CommandParams {
	return nil
}

// ParseColorHexCommand creates a new `/color {input}` command
func ParseColorHexCommand(input string) (*ColorCommand, error) {
	cmd := ColorCommand{
		input: input,
	}

	color, err := colorful.Hex(input)
	if err != nil {
		return nil, err
	}

	cmd.color = color

	return &cmd, nil
}

// ColorCommandResult is named the same in GraphQL
type ColorCommandResult struct {
	color colorful.Color
}

// Result resolved
func (cmd *ColorCommand) Result() *ColorCommandResult {
	srgb := ColorCommandResult{cmd.color}
	return &srgb
}

// ColorCommandRGB is named RGBColor in GraphQL
type ColorCommandRGB struct {
	color colorful.Color
}

// SRGB resolved
func (result *ColorCommandResult) SRGB() *ColorCommandRGB {
	srgb := ColorCommandRGB{result.color}
	return &srgb
}

// ColorSpaceName resolved
func (rgb *ColorCommandRGB) ColorSpaceName() string {
	return "sRGB"
}

// Hex resolved
func (rgb *ColorCommandRGB) Hex() string {
	return rgb.color.Hex()
}

// Red8Bit resolved
func (rgb *ColorCommandRGB) Red8Bit() int32 {
	red, _, _ := rgb.color.RGB255()
	return int32(red)
}

// Green8Bit resolved
func (rgb *ColorCommandRGB) Green8Bit() int32 {
	_, green, _ := rgb.color.RGB255()
	return int32(green)
}

// Blue8Bit resolved
func (rgb *ColorCommandRGB) Blue8Bit() int32 {
	_, _, blue := rgb.color.RGB255()
	return int32(blue)
}

// ColorCommandLab is named LabColor in GraphQL
type ColorCommandLab struct {
	l, a, b float64
}

// Lab resolved
func (result *ColorCommandResult) Lab() *ColorCommandLab {
	l, a, b := result.color.Lab()
	lab := ColorCommandLab{l, a, b}
	return &lab
}

// ColorSpaceName resolved
func (lab *ColorCommandLab) ColorSpaceName() string {
	return "Lab"
}

// L resolved
func (lab *ColorCommandLab) L() float64 {
	return lab.l
}

// A resolved
func (lab *ColorCommandLab) A() float64 {
	return lab.a
}

// B resolved
func (lab *ColorCommandLab) B() float64 {
	return lab.b
}

// Run converts the color to a preview
func (cmd *ColorCommand) Run(ctx context.Context) (CommandResult, error) {
	hex := cmd.color.Hex()

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<div style="width: 12em; height: 12em; background-color:` + hex + `"></div>`)
	htmlBuffer.WriteString(`<dl class="mt-4">`)
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">Hex</dt><dd>%s</dd>`, hex))
	red, green, blue := cmd.color.RGB255()
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">sRGB</dt><dd>rgb(%v, %v, %v)</dd>`, red, green, blue))
	l, a, b := cmd.color.Lab()
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">Lab</dt><dd>lab(%v %v %v)</dd>`, l, a, b))
	htmlBuffer.WriteString(`</dl>`)

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}

// A ColorGradientCommand represents the `/color gradient` command
type ColorGradientCommand struct {
	inputs []string
	colors []colorful.Color
}

// ParseColorGradientCommand creates a new `/color gradient` command
func ParseColorGradientCommand(params string) (*ColorGradientCommand, error) {
	lines := strings.Split(params, "\n")

	cmd := ColorGradientCommand{
		inputs: lines,
	}

	var colors []colorful.Color
	for _, line := range lines {
		color, err := colorful.Hex(line)
		if err != nil {
			return nil, err
		}
		colors = append(colors, color)
	}

	cmd.colors = colors

	return &cmd, nil
}

// Run converts the color to a preview
func (cmd *ColorGradientCommand) Run(ctx context.Context) (CommandResult, error) {
	var gradientStops []string
	for _, color := range cmd.colors {
		hex := color.Hex()
		gradientStops = append(gradientStops, hex)
	}

	cssLinearGradient := `linear-gradient(` + strings.Join(gradientStops, ", ") + `)`

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<div style="width: 12em; height: 12em; background: ` + cssLinearGradient + `"></div>`)
	htmlBuffer.WriteString(`<dl class="mt-4">`)
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">Hex</dt><dd>%s</dd>`, strings.Join(gradientStops, ", ")))
	htmlBuffer.WriteString(fmt.Sprintf(`<dt class="mt-2 font-bold">CSS Linear Gradient</dt><dd><code>%s</code></dd>`, cssLinearGradient))
	htmlBuffer.WriteString(`</dl>`)

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}
