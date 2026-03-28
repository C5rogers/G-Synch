package tests

import (
	"testing"

	"github.com/C5rogers/G-Synch/pkg/sync"
)

func TestLoader_StartStopAreSafe(t *testing.T) {
	loader := sync.NewLoader("testing loader")

	loader.Start()
	loader.Start()
	loader.Stop()
	loader.Stop()
}
