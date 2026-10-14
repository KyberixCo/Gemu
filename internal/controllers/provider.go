// Package controllers isolates model-specific APIs behind capability-driven providers.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrUnsupported = errors.New("controller feature unsupported")

type Capability struct {
	Supported bool   `json:"supported"`
	Reason    string `json:"reason,omitempty"`
}
type Battery struct {
	State   string `json:"state"`
	Level   string `json:"level,omitempty"`
	Percent *int   `json:"percent,omitempty"`
}
type Device struct {
	Provider     string                `json:"provider"`
	ID           string                `json:"id"` // Connection identifier, never treated as a physical UUID.
	StableID     string                `json:"stableId,omitempty"`
	Name         string                `json:"name"`
	Family       string                `json:"family"`
	PlayerNumber int                   `json:"playerNumber,omitempty"`
	Capabilities map[string]Capability `json:"capabilities"`
	Battery      Battery               `json:"battery"`
}
type FeatureRequest struct {
	Provider string          `json:"provider"`
	DeviceID string          `json:"deviceId"`
	Feature  string          `json:"feature"`
	Value    json.RawMessage `json:"value"`
}
type Provider interface {
	ID() string
	Devices(context.Context) ([]Device, error)
	SetFeature(context.Context, string, string, json.RawMessage) error
}
type Service struct{ providers []Provider }

func New(providers ...Provider) *Service { return &Service{providers: providers} }
func (s *Service) Devices(ctx context.Context) ([]Device, error) {
	result := []Device{}
	for _, provider := range s.providers {
		devices, err := provider.Devices(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", provider.ID(), err)
		}
		for _, device := range devices {
			device.Provider = provider.ID()
			result = append(result, device)
		}
	}
	return result, nil
}
func (s *Service) SetFeature(ctx context.Context, request FeatureRequest) error {
	for _, provider := range s.providers {
		if provider.ID() == request.Provider {
			devices, err := provider.Devices(ctx)
			if err != nil {
				return err
			}
			for _, device := range devices {
				if device.ID == request.DeviceID {
					if !device.Capabilities[request.Feature].Supported {
						return ErrUnsupported
					}
					return provider.SetFeature(ctx, device.ID, request.Feature, request.Value)
				}
			}
			return errors.New("controller disconnected")
		}
	}
	return ErrUnsupported
}
func DecodeXInputBattery(kind, level byte) Battery {
	switch kind {
	case 1:
		return Battery{State: "wired"}
	case 2, 3:
		levels := []string{"empty", "low", "medium", "full"}
		if int(level) < len(levels) {
			return Battery{State: "discharging", Level: levels[level]}
		}
	}
	return Battery{State: "unknown"}
}
