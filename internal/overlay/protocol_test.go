package overlay

import (
	"strings"
	"testing"
	"unicode/utf8"
	"unsafe"
)

func TestNativeProtocolLayout(t *testing.T) {
	if unsafe.Sizeof(Shared{}) != 560 || unsafe.Offsetof(Shared{}.Title) != 48 {
		t.Fatal("C++ IPC layout changed")
	}
}
func TestTitleTruncationPreservesUTF8AndTerminator(t *testing.T) {
	var title [512]byte
	setTitle(title[:], strings.Repeat("🎮 á", 200))
	value := strings.TrimRight(string(title[:]), "\x00")
	if !utf8.ValidString(value) || title[511] != 0 {
		t.Fatal("native title is not a terminated UTF-8 string")
	}
}
func TestProtocolStateAndLanguage(t *testing.T) {
	s := Shared{}
	s.configure("Game", "en", "running")
	if s.Language != 1 || s.State != 2 {
		t.Fatal(s)
	}
	s.configure("Different", "es", "closing")
	if s.Language != 0 || s.State != 3 || string(s.Title[:4]) != "Game" {
		t.Fatal("shared title must remain immutable after attachment")
	}
}
