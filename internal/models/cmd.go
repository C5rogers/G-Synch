package models

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
