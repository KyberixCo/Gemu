package assist

// The local control protocol intentionally has no arbitrary text/input endpoint.
// Input is produced by the user's controller or the floating tool itself.
type Request struct {
	Version int    `json:"version"`
	Action  string `json:"action"`
	HostPID uint32 `json:"hostPid,omitempty"`
	Session bool   `json:"session,omitempty"`
}
type Status struct {
	Version int    `json:"version"`
	Mode    Mode   `json:"mode"`
	Menu    uint64 `json:"menu"`
	Error   string `json:"error,omitempty"`
}

const protocolVersion = 1
