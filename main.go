package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/exp/slices"
)

type inputType int

const (
	inputTypeURL inputType = iota
	inputTypeMethod
	inputTypeHeader
	inputTypeData
)

type textInput struct {
	textinput.Model

	Type inputType
}

type model struct {
	ok     bool
	active int
	inputs []textInput
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			m.ok = true
			return m, tea.Quit

		case tea.KeyShiftTab, tea.KeyUp:
			m.prev()

		case tea.KeyTab, tea.KeyDown:
			m.next()

		case tea.KeyCtrlH:
			m.addHeader()

		case tea.KeyCtrlD:
			m.addData()

		case tea.KeyCtrlX:
			m.remove()

		case tea.KeyRunes:
			if m.inputs[m.active].Type == inputTypeMethod {
				upper(msg)
			}
		}
	}

	m.inputs[m.active].Model, cmd = m.inputs[m.active].Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s string

	s += "curl " + m.filteredInputs(inputTypeURL)[0].View() + "\n"
	s += "  -X " + m.filteredInputs(inputTypeMethod)[0].View() + "\n"
	for _, in := range m.filteredInputs(inputTypeHeader) {
		s += "  -H " + in.View() + "\n"
	}
	for _, in := range m.filteredInputs(inputTypeData) {
		s += "  -d " + in.View() + "\n"
	}

	s += "\n<ctrl-h> header | <ctrl-d> data | <ctrl-x> remove | <enter> build"

	return s + "\n"
}

func (m model) filteredInputs(t inputType) []textInput {
	var inputs []textInput
	for _, in := range m.inputs {
		if in.Type == t {
			inputs = append(inputs, in)
		}
	}
	return inputs
}

func (m *model) next() {
	m.active++
	if m.active > len(m.inputs)-1 {
		m.active = 0
	}
	m.focus()
}

func (m *model) prev() {
	m.active--
	if m.active < 0 {
		m.active = len(m.inputs) - 1
	}
	m.focus()
}

func (m model) focus() {
	for i := range m.inputs {
		if i == m.active {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *model) addHeader() {
	for i := len(m.inputs) - 1; i > 0; i-- {
		if m.inputs[i].Type <= inputTypeHeader {
			m.inputs = slices.Insert(m.inputs, i+1, newTextInput(inputTypeHeader, "Content-Type: application/json"))
			m.active = i + 1
			m.focus()
			return
		}
	}
}

func (m *model) addData() {
	for i := len(m.inputs) - 1; i > 0; i-- {
		if m.inputs[i].Type == inputTypeData || m.inputs[i].Type == inputTypeHeader {
			m.inputs = slices.Insert(m.inputs, i+1, newTextInput(inputTypeData, `{"foo":"bar"}`))
			m.active = i + 1
			m.focus()
			return
		}
	}
}

func (m *model) remove() {
	if m.active < 2 {
		return
	}

	m.inputs = append(m.inputs[:m.active], m.inputs[m.active+1:]...)

	if m.active > len(m.inputs)-1 {
		m.active = len(m.inputs) - 1
	}
	m.focus()
}

func upper(msg tea.KeyMsg) tea.KeyMsg {
	for i, r := range msg.Runes {
		if r >= 'a' && r <= 'z' {
			msg.Runes[i] = 'A' + (r - 'a')
		}
	}
	return msg
}

func initModel() model {
	m := model{}
	m.inputs = append(m.inputs, newTextInput(inputTypeURL, "https://www.example.com"))
	m.inputs = append(m.inputs, newTextInput(inputTypeMethod, "GET"))

	m.inputs[0].Focus()
	return m
}

func newTextInput(t inputType, placeholder string) textInput {
	in := textInput{Type: t, Model: textinput.New()}
	in.Prompt = ""
	in.Placeholder = placeholder
	return in
}

func build(m model) string {
	var (
		url     = m.filteredInputs(inputTypeURL)[0]
		method  = m.filteredInputs(inputTypeMethod)[0]
		headers = m.filteredInputs(inputTypeHeader)
		data    = m.filteredInputs(inputTypeData)
	)

	var cmd string
	cmd += "curl " + url.Value()
	if method.Value() != "" {
		cmd += " -X " + method.Value()
	}
	for _, h := range headers {
		cmd += " -H '" + h.Value() + "'"
	}
	for _, d := range data {
		cmd += " -d '" + d.Value() + "'"
	}
	return cmd
}

func pastecmd(s string) {
	cbs, err := syscall.ByteSliceFromString(s)
	if err != nil {
		panic(err)
	}
	for _, c := range cbs {
		syscall.RawSyscall(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSTI, uintptr(unsafe.Pointer(&c)))
	}
	fmt.Print("\r \r")
	os.Exit(0)
}

func main() {
	r, err := tea.NewProgram(initModel()).StartReturningModel()
	if err != nil {
		panic(err)
	}
	if m := r.(model); m.ok {
		fmt.Print("\n")
		pastecmd(build(m))
	}
}
