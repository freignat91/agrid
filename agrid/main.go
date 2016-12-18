package main

import ()

// build vars
var (
	Version       string
	Build         string
	clientManager = &ClientManager{}
	config        = &ClientConfig{}
)

func main() {
	config.init(Version, Build)
	cli()
}
