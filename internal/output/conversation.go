package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

const (
	colorBrand = "\033[38;2;232;15;19m"
	colorBold  = "\033[1m"
	colorDim   = "\033[2m"
)

// brandMark is the Interactive AI mark in half blocks, five rows tall.
var brandMark = []string{
	"  ██▄▄██  ",
	"▄▄  ██  ▄▄",
	"██  ██  ██",
	"▀▀  ██  ▀▀",
	"  ██▀▀██  ",
}

// ChatBanner is the header of a chat: the mark beside the wordmark and what to do.
func ChatBanner(agentName, sessionID string) string {
	text := []string{
		colorBold + "interactive AI" + colorReset,
		"",
		"Chatting with " + colorBold + agentName + colorReset,
		colorDim + "Session " + sessionID + " · resume with --session-id " + sessionID + colorReset,
		colorDim + "Ctrl+C or Ctrl+D to leave" + colorReset,
	}
	var b strings.Builder
	for i, mark := range brandMark {
		b.WriteString(colorBrand + mark + colorReset)
		if text[i] != "" {
			b.WriteString("  " + text[i])
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ChatPrompt is where the user types the next message on a terminal.
func ChatPrompt() string {
	return "\n" + colorBrand + colorBold + "›" + colorReset + " "
}

// ConversationPrinter writes a followed conversation. On a terminal each message is a
// labelled block and a message the agent is still writing is redrawn in place;
// otherwise each complete event is one plain line.
type ConversationPrinter struct {
	out    io.Writer
	tty    bool
	width  int
	labels map[string]string
	frames *FrameWriter
}

// NewConversationPrinter labels each event source by labels, or by its own name when missing.
func NewConversationPrinter(out io.Writer, labels map[string]string) *ConversationPrinter {
	return &ConversationPrinter{
		out:    out,
		tty:    IsTerminal(out),
		width:  terminalWidth(out),
		labels: labels,
	}
}

func (p *ConversationPrinter) TTY() bool {
	return p.tty
}

// Print shows an event; streaming marks a message the agent is still writing.
func (p *ConversationPrinter) Print(e deployment.AgentEvent, streaming bool) {
	if !p.tty {
		if !streaming {
			fmt.Fprintln(p.out, EventLine(e, p.labels))
		}
		return
	}
	block := EventBlock(e, p.labels, p.width)
	if !streaming && p.frames == nil {
		fmt.Fprint(p.out, block)
		return
	}
	if p.frames == nil {
		p.frames = NewFrameWriter(p.out)
	}
	p.frames.Write(block)
	if !streaming {
		p.frames = nil
	}
}

// Notice shows a transient state, such as the agent thinking, until the next event replaces it.
func (p *ConversationPrinter) Notice(text string) {
	if !p.tty {
		return
	}
	if p.frames == nil {
		p.frames = NewFrameWriter(p.out)
	}
	p.frames.Write("\n  " + colorDim + text + colorReset + "\n")
}

// Clear removes a transient notice.
func (p *ConversationPrinter) Clear() {
	if p.frames == nil {
		return
	}
	p.frames.Write("")
	p.frames = nil
}

// EventBlock renders an event for a terminal: a message as a labelled block wrapped
// to width, anything else as one dim line.
func EventBlock(e deployment.AgentEvent, labels map[string]string, width int) string {
	if e.Kind != deployment.EventKindMessage {
		line := "· " + e.Kind + " " + e.Source + ": " + EventSummary(e)
		if status := e.Status(); status != "" {
			line = "· " + status
		}
		// truncateString needs room for its "..." suffix.
		if width > 6 {
			line = truncateString(line, width-3)
		}
		return "  " + colorDim + line + colorReset + "\n"
	}

	label := sourceLabel(e.Source, labels)
	bullet := colorBlue
	if e.Source == deployment.EventSourceAgent {
		bullet = colorBrand
	}
	// One column short of the width, so a full line never wraps and breaks the redraw.
	body := EventSummary(e)
	if width > 3 {
		body = WrapText(body, width-3)
	}
	return "\n" + bullet + "●" + colorReset + " " + colorBold + label + colorReset + "\n" +
		"  " + strings.ReplaceAll(body, "\n", "\n  ") + "\n"
}

// WrapText breaks lines at spaces so none exceeds width, which keeps a redrawn
// frame's line count equal to the rows it occupies.
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		runes := []rune(line)
		for len(runes) > width {
			cut := width
			for i := width; i > 0; i-- {
				if runes[i] == ' ' {
					cut = i
					break
				}
			}
			out = append(out, strings.TrimRight(string(runes[:cut]), " "))
			runes = []rune(strings.TrimLeft(string(runes[cut:]), " "))
		}
		out = append(out, string(runes))
	}
	return strings.Join(out, "\n")
}

// EventFollower renders followed events. Structured output writes a record per
// complete event; the default shows them as a conversation, updated in place.
func EventFollower(out io.Writer, asJSON, asYAML bool) func(deployment.AgentEvent, bool) error {
	switch {
	case asJSON:
		return func(ev deployment.AgentEvent, streaming bool) error {
			if streaming {
				return nil
			}
			_, err := out.Write(append(ev.Raw, '\n'))
			return err
		}
	case asYAML:
		return func(ev deployment.AgentEvent, streaming bool) error {
			if streaming {
				return nil
			}
			if _, err := io.WriteString(out, "---\n"); err != nil {
				return err
			}
			return PrintRawYAML(out, ev.Raw)
		}
	}
	printer := NewConversationPrinter(out, nil)
	return func(ev deployment.AgentEvent, streaming bool) error {
		printer.Print(ev, streaming)
		return nil
	}
}

// sourceLabel names a message's sender by labels, or by its source when missing.
func sourceLabel(source string, labels map[string]string) string {
	if name, ok := labels[source]; ok {
		return name
	}
	return source
}
