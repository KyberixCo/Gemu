package overlay

import (
	"sync/atomic"
	"unicode/utf8"
)

const Version = 1
const (
	CommandNone uint32 = iota
	CommandResume
	CommandHome
	CommandClose
	CommandAssist
)

// Shared is also defined in native/overlay/protocol.h. All mutable controls
// are aligned uint32 values, exchanged atomically between the two processes.
type Shared struct {
	Version, Size                               uint32
	Ready, Visible, Command, State, Language    uint32
	HostTick, RendererTick, Stop, Width, Height uint32
	Title                                       [512]byte
}

func setTitle(target []byte, title string) {
	for len(title) >= len(target) {
		_, size := utf8.DecodeLastRuneInString(title)
		title = title[:len(title)-size]
	}
	clear(target)
	copy(target, title)
}

func stateCode(state string) uint32 {
	switch state {
	case "requested":
		return 1
	case "running":
		return 2
	case "closing":
		return 3
	}
	return 0
}

func (s *Shared) configure(title, language, state string) {
	// Title is written once, before loading the DLL.
	if s.Title[0] == 0 {
		setTitle(s.Title[:], title)
	}
	lang := uint32(0)
	if language == "en" {
		lang = 1
	}
	atomic.StoreUint32(&s.Language, lang)
	atomic.StoreUint32(&s.State, stateCode(state))
}
