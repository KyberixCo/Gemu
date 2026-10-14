package main

import (
	"context"
	"kyberix.com/gemu/internal/controllers"
)

// Shared across Gemu users: hardware capabilities do not belong to a library session.
var controllerService = controllers.New(controllers.Native())

func (a *App) GetControllers() ([]controllers.Device, error) {
	return controllerService.Devices(context.Background())
}
func (a *App) SetControllerFeature(request controllers.FeatureRequest) error {
	return controllerService.SetFeature(context.Background(), request)
}
