package sync

import (
	"testing"
	"time"
)

func TestLoaderStartIsIdempotent(t *testing.T) {
	originalSupports := supportsLoaderOutput
	supportsLoaderOutput = func() bool { return true }
	defer func() {
		supportsLoaderOutput = originalSupports
		loaderState.mu.Lock()
		loaderState.active = nil
		loaderState.mu.Unlock()
	}()

	loader := NewLoader("testing")
	loader.Start()
	loader.Start()
	loader.Stop()
	loader.Stop()

	select {
	case <-loader.doneCh:
	case <-time.After(time.Second):
		t.Fatal("loader did not stop")
	}
}

func TestLoaderStartStopsPreviousActiveLoader(t *testing.T) {
	originalSupports := supportsLoaderOutput
	supportsLoaderOutput = func() bool { return true }
	defer func() {
		supportsLoaderOutput = originalSupports
		loaderState.mu.Lock()
		loaderState.active = nil
		loaderState.mu.Unlock()
	}()

	first := NewLoader("first")
	second := NewLoader("second")

	first.Start()
	second.Start()
	defer second.Stop()

	select {
	case <-first.doneCh:
	case <-time.After(time.Second):
		t.Fatal("previous loader was not stopped")
	}

	loaderState.mu.Lock()
	active := loaderState.active
	loaderState.mu.Unlock()

	if active != second {
		t.Fatal("second loader should be active")
	}
}
