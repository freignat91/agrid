package main

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

type ClientManager struct {
	printColor [6]*color.Color
	server     string
	verbose    bool
	silence    bool
	debug      bool
}

var currentColorTheme = "default"
var (
	colRegular = 0
	colInfo    = 1
	colWarn    = 2
	colError   = 3
	colSuccess = 4
	colDebug   = 5
)

func (m *ClientManager) init() error {
	m.setColors()
	//
	return nil
}

func (m *ClientManager) getClient() (*gnodeClient, error) {
	client := gnodeClient{}
	err := client.init(m)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (m *ClientManager) printf(col int, format string, args ...interface{}) {
	if m.silence {
		return
	}
	colorp := m.printColor[0]
	if col > 0 && col < len(m.printColor) {
		colorp = m.printColor[col]
	}
	if !m.verbose && col == colInfo {
		return
	}
	if !m.debug && col == colDebug {
		return
	}
	colorp.Printf(format, args...)
}

func (m *ClientManager) Fatal(format string, args ...interface{}) {
	m.printf(colError, format, args...)
	os.Exit(1)
}

func (m *ClientManager) pError(format string, args ...interface{}) {
	m.printf(colError, format, args...)
}

func (m *ClientManager) pWarn(format string, args ...interface{}) {
	m.printf(colWarn, format, args...)
}

func (m *ClientManager) pInfo(format string, args ...interface{}) {
	m.printf(colInfo, format, args...)
}

func (m *ClientManager) pSuccess(format string, args ...interface{}) {
	m.printf(colSuccess, format, args...)
}

func (m *ClientManager) pRegular(format string, args ...interface{}) {
	m.printf(colRegular, format, args...)
}

func (m *ClientManager) pDebug(format string, args ...interface{}) {
	m.printf(colDebug, format, args...)
}

func (m *ClientManager) setColors() {
	theme := config.colorTheme
	if theme == "dark" {
		m.printColor[0] = color.New(color.FgHiWhite)
		m.printColor[1] = color.New(color.FgHiBlack)
		m.printColor[2] = color.New(color.FgYellow)
		m.printColor[3] = color.New(color.FgRed)
		m.printColor[4] = color.New(color.FgGreen)
		m.printColor[5] = color.New(color.FgHiBlack)
	} else {
		m.printColor[0] = color.New(color.FgMagenta)
		m.printColor[1] = color.New(color.FgHiBlack)
		m.printColor[2] = color.New(color.FgYellow)
		m.printColor[3] = color.New(color.FgRed)
		m.printColor[4] = color.New(color.FgGreen)
		m.printColor[5] = color.New(color.FgHiBlack)
	}
	//add theme as you want.
}

func (m *ClientManager) formatKey(key string) string {
	if key != "" {
		for len(key) < 32 {
			key = fmt.Sprintf("%s%s", key, key)
		}
		key = key[0:32]
	}
	return key
}
