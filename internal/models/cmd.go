package models

import (
	"github.com/fatih/color"
)

type CMD string

const (
	CHECK         CMD = "check"
	SYNCH         CMD = "synch"
	REVERSE_CHECK CMD = "reverse-check"
)

var CMDMapper = map[CMD]string{
	CHECK:         "check",
	SYNCH:         "synch",
	REVERSE_CHECK: "reverse-check",
}

type CheckReturn struct {
	Message string
	Type    string
	Label   string
}

func (c CheckReturn) GetColoredMessage() string {
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	switch c.Label {
	case "WARNING":
		return yellow(c.Message)
	case "ERROR":
		return red(c.Message)
	case "INFO":
		return blue(c.Message)
	case "SUCCESS":
		return green(c.Message)
	default:
		return c.Message
	}
}
