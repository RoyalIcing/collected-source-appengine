package main

type CommandParams struct {
}

// JSONEncoded resolved
func (params *CommandParams) JSONEncoded() *string {
	return nil
}

type Commands struct{}

// Color resolved
func (params *Commands) Color(args struct{ Input string }) (*ColorCommand, error) {
	return ParseColorHexCommand(args.Input)
}
