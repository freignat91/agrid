package main

import ()

// build vars
var (
	Version  string
	Build    string
	agridCli = &agridCLI{}
	config   = &CliConfig{}
)

func main() {
	config.init(Version, Build)
	cli()
}
