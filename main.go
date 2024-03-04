package main

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	columnStyle  = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type status int

const (
	todo status = iota
	doing
	done
)

type Task struct {
	status      status
	title       string
	description string
}

func (t Task) FilterValue() string {
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}

type Model struct {
	focused status
	lists   []list.Model
	err     error
	loaded  bool
	quite   bool
}

func NewModel() *Model {
	return &Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			m.initLists(msg.Width, msg.Height)
			m.loaded = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.quite = true
			return m, tea.Quit
		case "left", "h":
			m.Next()
		case "right", "l":
			m.Prev()
		case "enter":
			return m, m.MoveToNext
		}
	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quite {
		return "bye!"
	}
	if m.loaded {
		todoView := m.lists[todo].View()
		inprogressView := m.lists[doing].View()
		doneView := m.lists[done].View()

		switch m.focused {
		case todo:
			return lipgloss.JoinHorizontal(lipgloss.Left,
				columnStyle.Render(focusedStyle.Render(todoView)), columnStyle.Render(inprogressView), columnStyle.Render(doneView))
		case doing:
			return lipgloss.JoinHorizontal(lipgloss.Left,
				columnStyle.Render(todoView), columnStyle.Render(focusedStyle.Render(inprogressView)), columnStyle.Render(doneView))
		case done:
			return lipgloss.JoinHorizontal(lipgloss.Left,
				columnStyle.Render(todoView), columnStyle.Render(inprogressView), columnStyle.Render(focusedStyle.Render(doneView)))
		}

		return lipgloss.JoinHorizontal(lipgloss.Left, todoView, inprogressView, doneView)
	}
	return "loading..."

}

func (m *Model) Next() {
	m.focused++
	if m.focused > done {
		m.focused = todo
	}
}

func (m *Model) Prev() {
	m.focused--
	if m.focused < todo {
		m.focused = done
	}
}
func (m *Model) MoveToNext() tea.Msg {
	selectedItem := m.lists[m.focused].SelectedItem()
	selectedTask := selectedItem.(Task)
	m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
	selectedTask.Next()
	m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items())-1, list.Item(selectedTask))

	return nil
}

func (t *Task) Next() {
	t.status++
	if t.status > done {
		t.status = todo
	}
}

func (m *Model) initLists(width, height int) {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/3, height-3)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}

	// init first list
	m.lists[todo].Title = "To do"
	m.lists[todo].SetItems([]list.Item{
		Task{
			status:      todo,
			title:       "Do coffe",
			description: "Make delicious coffee",
		},
		Task{
			status:      todo,
			title:       "Buy milk",
			description: "Buy white milk 3.2% fat",
		},
		Task{
			status:      todo,
			title:       "Buy bread",
			description: "White fresh bread",
		},
	})

	// init second list
	m.lists[doing].Title = "In progress"
	m.lists[doing].SetItems([]list.Item{
		Task{
			status:      doing,
			title:       "Clean house",
			description: "We will have a quest soon",
		},
		Task{
			status:      doing,
			title:       "feed dog",
			description: "doggo woof woof",
		},
	})

	// init third list
	m.lists[done].Title = "Done"
	m.lists[done].SetItems([]list.Item{
		Task{
			status:      done,
			title:       "feed rabbit",
			description: "Bunny bunny",
		},
	})
}
func main() {
	p := tea.NewProgram(NewModel())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
