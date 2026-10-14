//go:build windows

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sys/windows"
	"sync"
	"unsafe"
)

type xinput struct {
	mu             sync.Mutex
	dll            *windows.LazyDLL
	state, battery *windows.LazyProc
	connected      [4]bool
	generation     [4]uint64
}

func Native() Provider {
	dll := windows.NewLazySystemDLL("xinput1_4.dll")
	return &xinput{dll: dll, state: dll.NewProc("XInputGetState"), battery: dll.NewProc("XInputGetBatteryInformation")}
}
func (*xinput) ID() string { return "xinput" }
func (p *xinput) Devices(ctx context.Context) ([]Device, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := []Device{}
	if p.dll.Load() != nil || p.state.Find() != nil {
		return result, nil
	}
	for slot := 0; slot < 4; slot++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state [4]uint32 // Aligned 16-byte XINPUT_STATE buffer.
		status, _, _ := p.state.Call(uintptr(slot), uintptr(unsafe.Pointer(&state)))
		if status != 0 {
			p.connected[slot] = false
			continue
		}
		if !p.connected[slot] {
			p.generation[slot]++
			p.connected[slot] = true
		}
		battery := Battery{State: "unknown"}
		supported := p.battery.Find() == nil
		if supported {
			var information [2]byte
			status, _, _ = p.battery.Call(uintptr(slot), 0, uintptr(unsafe.Pointer(&information)))
			if status == 0 {
				battery = DecodeXInputBattery(information[0], information[1])
			}
		}
		result = append(result, Device{ID: fmt.Sprintf("slot-%d-session-%d", slot, p.generation[slot]), Name: fmt.Sprintf("Xbox / XInput #%d", slot+1), Family: "xbox", PlayerNumber: slot + 1, Battery: battery, Capabilities: map[string]Capability{
			"battery": {Supported: supported}, "identity": {Reason: "XInput exposes a slot, not a persistent hardware ID"}, "player-led": {Reason: "XInput has no public player LED control API"},
		}})
	}
	return result, nil
}
func (*xinput) SetFeature(context.Context, string, string, json.RawMessage) error {
	return ErrUnsupported
}
