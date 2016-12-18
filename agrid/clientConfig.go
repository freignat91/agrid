package main

import (
	"fmt"
	"os"
	"strconv"
)

//AgentConfig configuration parameters
type ClientConfig struct {
	serverAddress string
	colorTheme    string
}

//update conf instance with default value and environment variables
func (cfg *ClientConfig) init(version string, build string) {
	cfg.setDefault()
	cfg.loadConfigUsingEnvVariable()
	//cfg.displayConfig(version, build)
}

//Set default value of configuration
func (cfg *ClientConfig) setDefault() {
	cfg.serverAddress = "127.0.0.1:30103"
	cfg.colorTheme = "dark"
}

//Update config with env variables
func (cfg *ClientConfig) loadConfigUsingEnvVariable() {
	cfg.serverAddress = cfg.getStringParameter("SERVER_ADDRESS", cfg.serverAddress)
	cfg.colorTheme = cfg.getStringParameter("COLOR_THEME", cfg.colorTheme)
}

//display amp-pilot configuration
func (cfg *ClientConfig) displayConfig(version string, build string) {
	fmt.Printf("agrid version: %v build: %s\n", version, build)
	fmt.Println("----------------------------------------------------------------------------")
	fmt.Println("Configuration:")
	fmt.Printf("agrid address: %s\n", cfg.serverAddress)
}

//return env variable value, if empty return default value
func (cfg *ClientConfig) getStringParameter(envVariableName string, def string) string {
	value := os.Getenv(envVariableName)
	if value == "" {
		return def
	}
	return value
}

//return env variable value convert to int, if empty return default value
func (cfg *ClientConfig) getIntParameter(envVariableName string, def int) int {
	value := os.Getenv(envVariableName)
	if value != "" {
		ivalue, err := strconv.Atoi(value)
		if err != nil {
			return def
		}
		return ivalue
	}
	return def
}
