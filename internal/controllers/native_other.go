//go:build !windows

package controllers

import (
	"context"
	"encoding/json"
)

type unavailable struct{}

func Native() Provider                                        { return unavailable{} }
func (unavailable) ID() string                                { return "native" }
func (unavailable) Devices(context.Context) ([]Device, error) { return []Device{}, nil }
func (unavailable) SetFeature(context.Context, string, string, json.RawMessage) error {
	return ErrUnsupported
}
