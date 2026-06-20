package tui

import (
	"context"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LogLineMsg is sent when a new log line is produced.
type LogLineMsg string

// ProcessFinishedMsg is sent when the process exits.
type ProcessFinishedMsg struct {
	Err error
}

// channelWriter redirects writes to a channel as LogLineMsg.
type channelWriter struct {
	ch  chan<- tea.Msg
	buf strings.Builder
}

func (cw *channelWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if b == '\n' {
			cw.ch <- LogLineMsg(cw.buf.String())
			cw.buf.Reset()
		} else {
			cw.buf.WriteByte(b)
		}
	}
	return len(p), nil
}

func (cw *channelWriter) Flush() {
	if cw.buf.Len() > 0 {
		cw.ch <- LogLineMsg(cw.buf.String())
		cw.buf.Reset()
	}
}

// Runner handles the background process execution.
type Runner struct {
	filePath    string
	cancel      context.CancelFunc
	ch          chan tea.Msg
	execCommand func(ctx context.Context, name string, arg ...string) *exec.Cmd
}

// NewRunner creates a new Runner for the given Go file path.
func NewRunner(filePath string) *Runner {
	return &Runner{
		filePath:    filePath,
		ch:          make(chan tea.Msg, 500),
		execCommand: exec.CommandContext,
	}
}

// Start spawns the workflow run in a background goroutine.
func (r *Runner) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	go func() {
		defer close(r.ch)

		cmd := r.execCommand(runCtx, "go", "run", r.filePath)

		cw := &channelWriter{ch: r.ch}
		cmd.Stdout = cw
		cmd.Stderr = cw

		err := cmd.Run()
		cw.Flush()

		r.ch <- ProcessFinishedMsg{Err: err}
	}()
}

// Stop terminates the background process if it is running.
func (r *Runner) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

// Channel returns the channel where process events are published.
func (r *Runner) Channel() <-chan tea.Msg {
	return r.ch
}

// WaitForActivity waits for an event from the runner's channel.
func WaitForActivity(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}
