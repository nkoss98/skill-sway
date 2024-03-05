package main

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
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

const (
	final status = iota
	form
)

var models []tea.Model

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
		case "n":
			models[final] = m // save the current model
			models[form] = NewForm(m.focused)
			return models[form].Update(nil)
		}
	case Task:
		task := msg
		return m, m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), list.Item(msg))
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

func NewForm(focused status) *Form {
	f := &Form{
		focused: focused,
	}
	f.title.Placeholder = "Title"
	f.title = textinput.New()
	f.title.Focus()
	f.description.Placeholder = "Description"
	f.description = textinput.New()

	return f
}

type Form struct {
	title       textinput.Model
	description textinput.Model
	focused     status
}

func (f Form) Init() tea.Cmd {
	return nil
}

func (f Form) CreateTask() tea.Msg {
	task := NewTask(f.focused, f.title.Value(), f.description.Value())
	return task
}

func NewTask(status status, title, description string) Task {
	return Task{
		status:      status,
		title:       title,
		description: description,
	}

}
func (f Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return f, tea.Quit
		case "enter":
			if f.title.Focused() {
				f.title.Blur()
				f.description.Focus()
				return f, textarea.Blink
			} else {
				models[form] = f
				return models[final], f.CreateTask
			}
		}
	}
	var cmd tea.Cmd
	if f.title.Focused() {
		f.title, cmd = f.title.Update(msg)
		return f, cmd
	} else {
		f.description, cmd = f.description.Update(msg)
		return f, cmd
	}
}

func (f Form) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, f.title.View(), f.description.View())
}

func main() {
	components := []tea.Model{
		NewModel(), NewForm(todo),
	}
	m := components[final]
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
