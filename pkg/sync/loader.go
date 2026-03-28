package sync

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

type Loader struct {
	message string
	stopCh  chan struct{}
	doneCh  chan struct{}
	mu      sync.Mutex
	started bool
	stopped bool
}

var loaderState struct {
	mu       sync.Mutex
	stdoutMu sync.Mutex
	active   *Loader
}

func NewLoader(message string) *Loader {
	return &Loader{
		message: message,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

func (l *Loader) Start() {
	if l == nil || !supportsLoaderOutput() {
		return
	}

	l.mu.Lock()
	if l.started {
		l.mu.Unlock()
		return
	}
	l.started = true
	l.mu.Unlock()

	loaderState.mu.Lock()
	previous := loaderState.active
	loaderState.active = l
	loaderState.mu.Unlock()

	if previous != nil && previous != l {
		previous.Stop()
	}

	go l.run()
}

func (l *Loader) Stop() {
	if l == nil || !supportsLoaderOutput() {
		return
	}

	l.mu.Lock()
	if !l.started {
		l.mu.Unlock()
		return
	}
	if !l.stopped {
		l.stopped = true
		close(l.stopCh)
	}
	doneCh := l.doneCh
	l.mu.Unlock()

	<-doneCh

	loaderState.mu.Lock()
	if loaderState.active == l {
		loaderState.active = nil
	}
	loaderState.mu.Unlock()
}

func (l *Loader) run() {
	defer close(l.doneCh)

	frames := []string{"|", "/", "-", "\\"}
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	frameIdx := 0
	for {
		select {
		case <-l.stopCh:
			clearLoaderLine()
			return
		case <-ticker.C:
			loaderState.stdoutMu.Lock()
			fmt.Fprintf(os.Stdout, "\r\033[K%s %s", frames[frameIdx], l.message)
			loaderState.stdoutMu.Unlock()
			frameIdx = (frameIdx + 1) % len(frames)
		}
	}
}

var supportsLoaderOutput = func() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func clearLoaderLine() {
	if !supportsLoaderOutput() {
		return
	}
	loaderState.stdoutMu.Lock()
	fmt.Fprint(os.Stdout, "\r\033[K")
	loaderState.stdoutMu.Unlock()
}

func pauseActiveLoader() {
	loaderState.mu.Lock()
	defer loaderState.mu.Unlock()

	if loaderState.active != nil {
		clearLoaderLine()
	}
}

func Printf(writer *bufio.Writer, format string, args ...interface{}) {
	if writer != nil {
		fmt.Fprintf(writer, format, args...)
		return
	}

	pauseActiveLoader()
	loaderState.stdoutMu.Lock()
	fmt.Printf(format, args...)
	loaderState.stdoutMu.Unlock()
}

func Println(writer *bufio.Writer, args ...interface{}) {
	if writer != nil {
		fmt.Fprintln(writer, args...)
		return
	}

	pauseActiveLoader()
	loaderState.stdoutMu.Lock()
	fmt.Println(args...)
	loaderState.stdoutMu.Unlock()
}
