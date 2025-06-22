package edit

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Add    key.Binding
	Remove key.Binding
	Help   key.Binding
	Reload key.Binding
	Select key.Binding
	Quit   key.Binding
	Undo   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // first column
		{k.Add, k.Remove, k.Select, k.Undo},
		{k.Reload, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h", "shift+tab"),
		key.WithHelp("h/shift+tab", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l", "tab"),
		key.WithHelp("l/tab", "move right"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add row"),
	),
	Remove: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d/del", "delete row"),
	),
	Select: key.NewBinding(
		key.WithKeys("e", "enter"),
		key.WithHelp("e/enter", "edit cell"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
	),
	Undo: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "undo"),
	),
}
