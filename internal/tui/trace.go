package tui

import (
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Tracer wraps a tea.Model and writes one line per message to w: how long
// the program waited for it, which kind of message it was, and how long
// Update and View took. It exists to tell "the program is slow" apart from
// "input reaches the program late" (terminal, SSH, tmux). Key text is never
// logged, because it can be a password.
type Tracer struct {
	inner   tea.Model
	w       io.Writer
	start   time.Time
	last    time.Time
	started bool
}

// NewTracer wraps inner and writes a header with non-sensitive terminal
// facts (TERM, whether SSH/tmux/screen are in use) to w.
func NewTracer(inner tea.Model, w io.Writer) *Tracer {
	now := time.Now()
	t := &Tracer{inner: inner, w: w, start: now, last: now}
	_, _ = fmt.Fprintf(w, "# unattend-gen trace; TERM=%q COLORTERM=%q ssh=%t tmux=%t screen=%t\n",
		os.Getenv("TERM"), os.Getenv("COLORTERM"),
		os.Getenv("SSH_CONNECTION") != "", os.Getenv("TMUX") != "", os.Getenv("STY") != "")
	if fi, err := os.Stat("/proc/self"); err == nil {
		_, _ = fmt.Fprintf(w, "# process start -> tracing start: %v (includes the terminal colour query)\n",
			now.Sub(fi.ModTime()).Round(time.Millisecond))
	}
	return t
}

// Init forwards to the wrapped model.
func (t *Tracer) Init() tea.Cmd { return t.inner.Init() }

// Update forwards msg and logs timings.
func (t *Tracer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	now := time.Now()
	gap := now.Sub(t.last)
	next, cmd := t.inner.Update(msg)
	t.inner = next
	upd := time.Since(now)
	t.last = time.Now()
	kind := fmt.Sprintf("%T", msg)
	if _, ok := msg.(tea.KeyMsg); ok {
		kind = "tea.KeyMsg" // never log which key
	}
	_, _ = fmt.Fprintf(t.w, "t=%-9v waited=%-9v update=%-9v %s\n",
		now.Sub(t.start).Round(time.Millisecond), gap.Round(time.Millisecond), upd.Round(time.Microsecond), kind)
	return t, cmd
}

// View forwards to the wrapped model and logs slow or first renders.
func (t *Tracer) View() string {
	s := time.Now()
	out := t.inner.View()
	if d := time.Since(s); !t.started || d > 5*time.Millisecond {
		_, _ = fmt.Fprintf(t.w, "t=%-9v view=%v (first=%t)\n", s.Sub(t.start).Round(time.Millisecond), d.Round(time.Microsecond), !t.started)
	}
	t.started = true
	return out
}
