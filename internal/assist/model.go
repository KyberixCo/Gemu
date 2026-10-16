// Package assist implements the desktop input companion independently of the
// Gemu window, game session, WebView and selected player.
package assist

import (
	"math"
	"time"
)

type Mode string

const (
	Off      Mode = "off"
	Pointer  Mode = "pointer"
	Keyboard Mode = "keyboard"
)

const (
	Up    uint16 = 0x1
	Down  uint16 = 0x2
	Left  uint16 = 0x4
	Right uint16 = 0x8
	Start uint16 = 0x10
	Back  uint16 = 0x20
	R3    uint16 = 0x80
	LB    uint16 = 0x100
	RB    uint16 = 0x200
	Guide uint16 = 0x400
	A     uint16 = 0x1000
	B     uint16 = 0x2000
	X     uint16 = 0x4000
	Y     uint16 = 0x8000
)

// Shortcuts delays the standalone menu action until the base chord is released.
// This gives Home+X and Home+R3 a chance to consume it, in either press order.
type Shortcuts struct {
	ready, base, consumed bool
}

func (s *Shortcuts) Reset() { *s = Shortcuts{} }
func (s *Shortcuts) Sample(buttons uint16) (mode Mode, menu, suppress bool) {
	if !s.ready {
		if buttons == 0 {
			s.ready = true
		}
		return "", false, true
	}
	base := buttons&Guide != 0 || buttons&(Back|Start) == Back|Start
	if base {
		s.base = true
		if !s.consumed {
			if buttons&X != 0 {
				mode = Keyboard
				s.consumed = true
			} else if buttons&R3 != 0 {
				mode = Pointer
				s.consumed = true
			}
		}
		return mode, false, true
	}
	if s.base {
		menu = !s.consumed
		if menu {
			s.consumed = true
		}
		// Consume the entire chord, including X/R3 released after Home.
		if buttons == 0 {
			s.base = false
			s.consumed = false
		}
		return "", menu, true
	}
	return "", false, false
}

func velocity(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) <= .18 {
		return 0
	}
	return math.Copysign(math.Pow((math.Min(1, math.Abs(value))-.18)/.82, 2), value)
}

type Key struct{ Label, Text, Command string }

func keyRows(symbols, upper bool) [][]Key {
	rows := []string{"1234567890", "qwertyuiop", "asdfghjklñ", "zxcvbnm,.-"}
	if upper {
		rows = []string{"1234567890", "QWERTYUIOP", "ASDFGHJKLÑ", "ZXCVBNM,.-"}
	}
	if symbols {
		rows = []string{"!@#$%&*()_", "+=[]{}<>^~", "`|\\/;:'\"?", "áéíóúüÁÉÍÓÚÜ"}
	}
	result := make([][]Key, 0, 6)
	for _, row := range rows {
		keys := []Key{}
		for _, char := range row {
			keys = append(keys, Key{Label: string(char), Text: string(char)})
		}
		result = append(result, keys)
	}
	result = append(result, []Key{{Label: "Mayús", Command: "shift"}, {Label: "#+=", Command: "symbols"}, {Label: "Espacio", Text: " "}, {Label: "Borrar", Command: "backspace"}, {Label: "Tab", Command: "tab"}, {Label: "Enter", Command: "enter"}})
	return append(result, []Key{{Label: "←", Command: "left"}, {Label: "→", Command: "right"}, {Label: "↑", Command: "up"}, {Label: "↓", Command: "down"}, {Label: "Esc", Command: "escape"}, {Label: "Mover", Command: "move"}, {Label: "Ratón", Command: "pointer"}, {Label: "Listo", Command: "close"}})
}

type navigationRepeat struct {
	held uint16
	next time.Time
}

func (r *navigationRepeat) sample(direction uint16, now time.Time) bool {
	if direction == 0 {
		r.held = 0
		return false
	}
	if direction != r.held {
		r.held = direction
		r.next = now.Add(350 * time.Millisecond)
		return true
	}
	if now.Before(r.next) {
		return false
	}
	r.next = now.Add(100 * time.Millisecond)
	return true
}
