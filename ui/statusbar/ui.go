package statusbar

import (
	"fmt"
	"github.com/blackmann/go-gurl/lib"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommandInput is an entry from the keyboard
// that is mapped to an action. For example, ":q" to quit, lib.
// The action is not necessarily performed by statusbar. We're
// only using this type as a state update message type.
//
//  statusbar.Update(commandInput(":q"))
//
// This is a tea.Msg type
type CommandInput string

type Model struct {
	spinner  spinner.Model
	spinning bool

	width        int
	status       lib.Status
	commandEntry string
	message      lib.ShortMessage
	mode         lib.Mode
}

func NewStatusBar() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{spinner: s}
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CommandInput:
		model.commandEntry = string(msg)
		return model, nil

	case lib.Status:
		// Allow to flow through so ticking can begin for
		// status == PROCESSING
		model.status = msg

	case tea.WindowSizeMsg:
		model.width = msg.Width
		return model, nil

	case lib.ShortMessage:
		model.message = msg
		return model, nil

	case lib.Mode:
		model.mode = msg
		return model, nil
	}

	if model.status == lib.PROCESSING {
		if !model.spinning {
			model.spinning = true
			return model, model.spinner.Tick
		}

		var cmd tea.Cmd
		model.spinner, cmd = model.spinner.Update(msg)
		return model, cmd
	} else {
		model.spinning = false
	}

	return model, nil
}

func (model Model) View() string {
	var status string

	switch model.status {
	case lib.PROCESSING:
		status = fmt.Sprintf("%s Processing", model.spinner.View())
	case lib.IDLE:
		status = neutralStatusStyle.Render("Idle")

	case lib.ERROR:
		status = errorStatusStyle.Render("Error")

	default:
		value := model.status.GetValue()

		if value < 300 {
			status = okStatusStyle.Render(fmt.Sprintf("%d", value))
		} else if value < 400 {
			status = okStatusStyle.Render(fmt.Sprintf("%d", value))
		} else {
			status = errorStatusStyle.Render(fmt.Sprintf("%d", value))
		}
	}

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#999"))

	halfWidth := model.width/2 - 2
	half := lipgloss.NewStyle().Width(halfWidth) // Left/right padding removed

	// we need to truncate the command entry to fit one line
	commandEntry := model.commandEntry
	modeLength := len(string(model.mode)) + 2 // for space and colon

	// TODO: Move this to utils
	truncateIndex := len(commandEntry) - (halfWidth - modeLength)
	if halfWidth > 0 && truncateIndex > 0 {
		commandEntry = fmt.Sprintf("> …%s", commandEntry[truncateIndex+3:])
	}

	rightHalf := half.Copy().Align(lipgloss.Right).
		MaxHeight(1).
		Render(fmt.Sprintf("%s :%s", commandEntry, mutedStyle.Render(string(model.mode))))

	leftHalf := half.Copy().Align(lipgloss.Left).
		Render(fmt.Sprintf("%s %s", status, mutedStyle.Render(string(model.message))))

	render := fmt.Sprintf("%s %s", leftHalf, rightHalf)

	return barStyle.Copy().Width(model.width).Render(render)
}
