package main

import (
	"github.com/fatih/color"
	"os"
)

type agridCLI struct {
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

func (m *agridCLI) init() error {
	m.setColors()
	//
	return nil
}

func (m *agridCLI) printf(col int, format string, args ...interface{}) {
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

func (m *agridCLI) Fatal(format string, args ...interface{}) {
	m.printf(colError, format, args...)
	os.Exit(1)
}

func (m *agridCLI) pError(format string, args ...interface{}) {
	m.printf(colError, format, args...)
}

func (m *agridCLI) pWarn(format string, args ...interface{}) {
	m.printf(colWarn, format, args...)
}

func (m *agridCLI) pInfo(format string, args ...interface{}) {
	m.printf(colInfo, format, args...)
}

func (m *agridCLI) pSuccess(format string, args ...interface{}) {
	m.printf(colSuccess, format, args...)
}

func (m *agridCLI) pRegular(format string, args ...interface{}) {
	m.printf(colRegular, format, args...)
}

func (m *agridCLI) pDebug(format string, args ...interface{}) {
	m.printf(colDebug, format, args...)
}

func (m *agridCLI) setColors() {
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
