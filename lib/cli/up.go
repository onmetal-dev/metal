package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewModel is an initializer which creates a new model for rendering our Bubbletea app.
func NewModel() (*model, error) {
	// We need to initialize a new text input model.
	ti := textinput.New()
	ti.CharLimit = 30
	ti.Focus()
	ti.Placeholder = "Type in your event"
	// Nest the text input in our application state.
	return &model{input: ti}, nil
}

type model struct {
	// nameInput stores the event name we have from the text input component
	nameInput string
	// listinput stores the event name selected from the list, used as an
	// autocomplete.
	listInput string
	// event stores the final selected event.
	event string

	input textinput.Model
}

// Ensure that model fulfils the tea.Model interface at compile time.
var _ tea.Model = (*model)(nil)

// Init() is called to kick off the render cycle.  It allows you to
// perform IO after the app has loaded and rendered once, asynchronously.
// The tea.Cmd can return a tea.Msg which will be passed into Update() in order
// to update the model's state.
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update is called with a tea.Msg, representing something that happened within
// our application.
//
// This can be things like terminal resizing, keypresses, or custom IO.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Let's figure out what is in tea.Msg, and what we need to do.
	switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	// 	// The terminal was resized.  We can access the new size with:
	// 	_, _ = msg.Width, msg.Height
	case tea.KeyMsg:
		// msg is a keypress.  We can handle each key combo uniquely, and update
		// our state:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlBackslash:
			// In this case, ctrl+c or ctrl+backslash quits the app by sending a
			// tea.Quit cmd.  This is a Bubbletea builtin which terminates the
			// overall framework which renders our model.
			//
			// Unfortunately, if you don't include this quitting can be, uh,
			// frustrating, as bubbletea catches every key combo by default.
			return m, tea.Quit
		case tea.KeyEnter:
			m.event = m.input.Value()
			return m, tea.Quit
		}
	}
	// We return an updated model to Bubbletea for rendering here.  This allows
	// us to mutate state so that Bubbletea can render an updated view.
	//
	// We also return "commands".  A command is something that you need to do
	// after rendering.  Each command produces a tea.Msg which is its *result*.
	// Bubbletea calls this Update function again with the tea.Msg - this is our
	// render loop.

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders output to the CLI.
func (m model) View() string {
	if m.event != "" {
		return fmt.Sprintf("You've selected: %s", m.event)
	}

	b := &strings.Builder{}
	b.WriteString("Enter your event:\n")
	// render the text input.  All we need to do to show the full
	// input is call View() and return the string.
	b.WriteString(m.input.View())
	return b.String()
}

func NewCmdUp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "up",
		Short:   "Launch an application",
		Example: "metal up .",
		PreRun:  checkToken,
		Run:     runUp,
	}
	return cmd
}

func checkToken(cmd *cobra.Command, args []string) {
	token := viper.GetString("token")
	if token == "" {
		fmt.Println("Error: Token is not set. Please set it using --token flag or in the config file.")
		os.Exit(1)
	}
}

func runUp(cmd *cobra.Command, args []string) {
	// Create a new TUI model which will be rendered in Bubbletea.
	state, err := NewModel()
	if err != nil {
		fmt.Printf("Error starting init command: %s\n", err)
		os.Exit(1)
	}
	// tea.NewProgram starts the Bubbletea framework which will render our
	// application using our state.
	p := tea.NewProgram(state)
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Print the final state after the program exits
	fmt.Println(finalModel.(model).View())
}
