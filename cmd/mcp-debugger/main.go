package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/lightningnetwork/lnd/actor"
	"github.com/mattn/go-isatty"
	mcpdebug "github.com/roasbeef/mcp-debug"
)

// model defines the state of our TUI application.
type model struct {
	// choices are the different options the user can select.
	choices []string

	// cursor is the current choice the user has selected.
	cursor int

	// debugger is a reference to the debugger actor factory.
	debugger actor.ActorRef[*mcpdebug.DebuggerMsg, *mcpdebug.DebuggerResp]

	// sessionKey is the service key for the current debugging session.
	sessionKey actor.ServiceKey[*mcpdebug.DAPRequest, *mcpdebug.DAPResponse]

	// status is a message to display to the user.
	status string
}

// initialModel returns the initial state of our application.
func initialModel() model {
	return model{
		choices: []string{"Start Debugger", "Exit"},
	}
}

// Init is the first command that is run when the application starts.
func (m model) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received. It's responsible for
// updating the model and returning a new command.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter":
			switch m.choices[m.cursor] {
			case "Start Debugger":
				// Send a message to the debugger actor to start
				// a new session.
				startMsg := &mcpdebug.DebuggerMsg{
					Command: &mcpdebug.StartDebuggerCmd{},
				}
				future := m.debugger.Ask(context.Background(), startMsg)
				go func() {
					resp, err := future.Await(context.Background()).Unpack()
					if err != nil {
						m.status = fmt.Sprintf("Error: %v", err)
						return
					}
					m.status = fmt.Sprintf("New session started: %s", resp.Status)
					m.sessionKey = actor.NewServiceKey[*mcpdebug.DAPRequest, *mcpdebug.DAPResponse](resp.Status)

					// Get a reference to the session actor.
					sessionRefs := actor.FindInReceptionist(
						mcpdebug.System.Receptionist(), m.sessionKey,
					)
					if len(sessionRefs) == 0 {
						m.status = "Error: Could not find session actor"
						return
					}
					sessionRef := sessionRefs[0]

					// Initialize the session.
					initResp, err := mcpdebug.InitializeSession(sessionRef, "mcp-debug-tui")
					if err != nil {
						m.status = fmt.Sprintf("Error: %v", err)
						return
					}
					m.status = fmt.Sprintf("Session initialized: %+v", initResp)
				}()

			case "Exit":
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View is responsible for rendering the UI.
func (m model) View() string {
	var b strings.Builder

	b.WriteString("MCP Debugger\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s [%s]\n", cursor, choice))
	}

	b.WriteString(fmt.Sprintf("\nStatus: %s\n", m.status))
	b.WriteString("\nPress q to quit.\n")

	return b.String()
}

func main() {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println("This application requires an interactive terminal.")
		os.Exit(1)
	}

	mcpdebug.StartDaemon()

	debuggerRefs := actor.FindInReceptionist(
		mcpdebug.System.Receptionist(), mcpdebug.DebuggerKey,
	)
	if len(debuggerRefs) == 0 {
		fmt.Println("Error: Could not find debugger actor factory")
		os.Exit(1)
	}
	debuggerRef := debuggerRefs[0]

	m := initialModel()
	m.debugger = debuggerRef

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}