package retroarch

import (
	"context"
	"net/http"
	"sync"
	"time"

	"kyberix.com/gemu/internal/library"
)

type Firmware struct {
	Name   string `json:"name"`
	MD5    string `json:"md5"`
	Region string `json:"region"`
}
type System struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Core       string     `json:"core"`
	CoreName   string     `json:"coreName"`
	Extensions []string   `json:"extensions"`
	Firmware   []Firmware `json:"firmware"`
}

var Systems = []System{
	{ID: "nes", Name: "Nintendo NES", Core: "fceumm", CoreName: "FCEUmm", Extensions: []string{".nes", ".unf", ".unif"}},
	{ID: "snes", Name: "Super Nintendo", Core: "snes9x", CoreName: "Snes9x", Extensions: []string{".sfc", ".smc"}},
	{ID: "gb", Name: "Game Boy", Core: "gambatte", CoreName: "Gambatte", Extensions: []string{".gb"}},
	{ID: "gbc", Name: "Game Boy Color", Core: "gambatte", CoreName: "Gambatte", Extensions: []string{".gbc"}},
	{ID: "gba", Name: "Game Boy Advance", Core: "mgba", CoreName: "mGBA", Extensions: []string{".gba"}},
	{ID: "megadrive", Name: "Mega Drive / Genesis", Core: "genesis_plus_gx", CoreName: "Genesis Plus GX", Extensions: []string{".md", ".gen", ".bin", ".smd"}},
	{ID: "mastersystem", Name: "Master System", Core: "genesis_plus_gx", CoreName: "Genesis Plus GX", Extensions: []string{".sms"}},
	{ID: "gamegear", Name: "Game Gear", Core: "genesis_plus_gx", CoreName: "Genesis Plus GX", Extensions: []string{".gg"}},
	{ID: "n64", Name: "Nintendo 64", Core: "mupen64plus_next", CoreName: "Mupen64Plus-Next", Extensions: []string{".n64", ".z64", ".v64"}},
	{ID: "psx", Name: "PlayStation", Core: "mednafen_psx_hw", CoreName: "Beetle PSX HW", Extensions: []string{".cue", ".chd", ".pbp", ".m3u"}, Firmware: []Firmware{
		{Name: "scph5500.bin", MD5: "8dd7d5296a650fac7319bce665a6a53c", Region: "Japan"},
		{Name: "scph5501.bin", MD5: "490f666e1afb15b7362b406ed1cea246", Region: "USA"},
		{Name: "scph5502.bin", MD5: "32736f17079d0b2b7024407c39bd3050", Region: "Europe"},
	}},
}

type Collection struct {
	SystemID  string `json:"systemId"`
	Directory string `json:"directory"`
	Recursive bool   `json:"recursive"`
}
type Config struct {
	CoreDirectory   string       `json:"coreDirectory"`
	Executable      string       `json:"executable"`
	SystemDirectory string       `json:"systemDirectory"`
	Collections     []Collection `json:"collections"`
	Fullscreen      bool         `json:"fullscreen"`
	Rewind          bool         `json:"rewind"`
	Language        string       `json:"language"`
	JoypadDriver    string       `json:"joypadDriver"`
}
type Installation struct {
	Executable string `json:"executable"`
	Managed    bool   `json:"managed"`
}
type Status struct {
	Supported          bool              `json:"supported"`
	Installations      []Installation    `json:"installations"`
	Config             Config            `json:"config"`
	Systems            []System          `json:"systems"`
	Diagnostics        []string          `json:"diagnostics"`
	DefaultDirectories map[string]string `json:"defaultDirectories"`
}
type FirmwareStatus struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	State  string `json:"state"`
}
type Check struct {
	SystemID  string           `json:"systemId"`
	CorePath  string           `json:"corePath"`
	CoreReady bool             `json:"coreReady"`
	Firmware  []FirmwareStatus `json:"firmware"`
}
type Preview struct {
	Token       string          `json:"token"`
	Games       []library.Entry `json:"games"`
	Checks      []Check         `json:"checks"`
	Diagnostics []string        `json:"diagnostics"`
	CanPublish  bool            `json:"canPublish"`
}
type InstallRequest struct {
	CoreDirectory string   `json:"coreDirectory"`
	Executable    string   `json:"executable"`
	InstallApp    bool     `json:"installApp"`
	Systems       []string `json:"systems"`
}
type Download struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
}
type Plan struct {
	RecommendedBytes int64      `json:"recommendedBytes"`
	FreeBytes        int64      `json:"freeBytes"`
	Token            string     `json:"token"`
	Version          string     `json:"version"`
	Destination      string     `json:"destination"`
	Downloads        []Download `json:"downloads"`
}
type Job struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	Current    string `json:"current"`
	Bytes      int64  `json:"bytes"`
	Total      int64  `json:"total"`
	Completed  int    `json:"completed"`
	Count      int    `json:"count"`
	Executable string `json:"executable"`
	Error      string `json:"error"`
}
type installPlan struct {
	plan    Plan
	owner   string
	request InstallRequest
	expires time.Time
}
type installJob struct {
	view   Job
	owner  string
	cancel context.CancelFunc
}
type draft struct {
	baseline string
	owner    string
	config   Config
	preview  Preview
	expires  time.Time
}
type Service struct {
	dir    string
	client *http.Client
	mu     sync.Mutex
	plans  map[string]installPlan
	jobs   map[string]*installJob
	drafts map[string]draft
	busy   bool
	// Network and extraction seams allow deterministic tests without installing
	// or executing downloaded software on a developer's machine.
	extract func(context.Context, string, string, func(string)) error
	checkPE func(string, bool) error
}

func system(id string) (System, bool) {
	for _, s := range Systems {
		if s.ID == id {
			return s, true
		}
	}
	return System{}, false
}
