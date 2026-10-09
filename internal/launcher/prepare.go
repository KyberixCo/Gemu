package launcher

import (
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/providers/steam"
)

// PrepareSteam does not claim a running session: handing off a URI yields no game PID.
func PrepareSteam(installation domain.Installation) (domain.LaunchSpec, error) {
	if installation.Provider != "steam" || !installation.Installed || !steam.ValidAppID(installation.ExternalID) {
		return domain.LaunchSpec{}, fmt.Errorf("Steam installation unavailable or invalid")
	}
	return domain.LaunchSpec{URI: "steam://run/" + installation.ExternalID, Confidence: "low"}, nil
}
