package assist

import "testing"

func TestShortcutsConsumeWholeChord(t *testing.T) {
	for _, base := range []uint16{Guide, Back | Start} {
		for _, extra := range []uint16{X, R3} {
			for _, first := range []uint16{base, extra} {
				s := Shortcuts{}
				s.Sample(0)
				s.Sample(first)
				mode, menu, suppress := s.Sample(base | extra)
				want := Keyboard
				if extra == R3 {
					want = Pointer
				}
				if mode != want || menu || !suppress {
					t.Fatalf("chord %#x: %q %v %v", base|extra, mode, menu, suppress)
				}
				for _, buttons := range []uint16{base | extra, extra, extra, 0} {
					mode, menu, _ = s.Sample(buttons)
					if mode != "" || menu {
						t.Fatalf("chord leaked on release %#x", buttons)
					}
				}
			}
		}
	}
}
func TestMenuWaitsForReleaseAndRearms(t *testing.T) {
	s := Shortcuts{}
	s.Sample(0)
	for i := 0; i < 2; i++ {
		_, menu, _ := s.Sample(Guide)
		if menu {
			t.Fatal("menu fired before chord resolution")
		}
		_, menu, _ = s.Sample(0)
		if !menu {
			t.Fatal("missing released menu")
		}
		_, menu, _ = s.Sample(0)
		if menu {
			t.Fatal("menu repeated")
		}
	}
	s.Reset()
	mode, menu, _ := s.Sample(Guide | X)
	if mode != "" || menu {
		t.Fatal("held buttons activated after device/focus change")
	}
}
func TestKeyboardCanTypePrintableASCII(t *testing.T) {
	available := map[rune]bool{}
	for _, mode := range [][2]bool{{false, false}, {false, true}, {true, false}} {
		for _, row := range keyRows(mode[0], mode[1]) {
			for _, key := range row {
				for _, char := range key.Text {
					available[char] = true
				}
			}
		}
	}
	for char := rune(32); char <= 126; char++ {
		if !available[char] {
			t.Errorf("missing printable character %q", char)
		}
	}
}
