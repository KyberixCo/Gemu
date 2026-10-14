package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakeProvider struct {
	devices []Device
	calls   int
	err     error
}

func (*fakeProvider) ID() string                                  { return "test" }
func (p *fakeProvider) Devices(context.Context) ([]Device, error) { return p.devices, p.err }
func (p *fakeProvider) SetFeature(context.Context, string, string, json.RawMessage) error {
	p.calls++
	return nil
}
func TestFeatureCapabilities(t *testing.T) {
	p := &fakeProvider{devices: []Device{{ID: "a", Capabilities: map[string]Capability{"player-led": {Supported: true}}}}}
	service := New(p)
	if err := service.SetFeature(context.Background(), FeatureRequest{Provider: "test", DeviceID: "a", Feature: "battery"}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if p.calls != 0 {
		t.Fatal("unsupported feature reached provider")
	}
	if err := service.SetFeature(context.Background(), FeatureRequest{Provider: "test", DeviceID: "a", Feature: "player-led", Value: json.RawMessage(`2`)}); err != nil {
		t.Fatal(err)
	}
	if p.calls != 1 {
		t.Fatal("supported feature did not reach provider")
	}
	if err := service.SetFeature(context.Background(), FeatureRequest{Provider: "test", DeviceID: "gone", Feature: "player-led"}); err == nil {
		t.Fatal("accepted disconnected device")
	}
	if err := service.SetFeature(context.Background(), FeatureRequest{Provider: "missing"}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}
func TestDeviceProviders(t *testing.T) {
	p := &fakeProvider{devices: []Device{{ID: "a"}}}
	devices, err := New(p).Devices(context.Background())
	if err != nil || len(devices) != 1 || devices[0].Provider != "test" {
		t.Fatalf("%v %v", devices, err)
	}
	p.err = errors.New("failure")
	if _, err := New(p).Devices(context.Background()); err == nil {
		t.Fatal("lost provider error")
	}
}
func TestBatteryLevels(t *testing.T) {
	for level, name := range []string{"empty", "low", "medium", "full"} {
		battery := DecodeXInputBattery(2, byte(level))
		if battery.Level != name || battery.Percent != nil {
			t.Fatalf("%+v", battery)
		}
	}
	if DecodeXInputBattery(1, 3).State != "wired" {
		t.Fatal("wired battery misreported")
	}
	for _, kind := range []byte{0, 255, 7} {
		if DecodeXInputBattery(kind, 3).State != "unknown" {
			t.Fatal("invented battery level")
		}
	}
	if DecodeXInputBattery(3, 99).State != "unknown" {
		t.Fatal("invalid level accepted")
	}
}
