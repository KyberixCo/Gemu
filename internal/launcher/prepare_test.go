package launcher

import (
	"kyberix.com/gemu/internal/domain"
	"testing"
)

func TestPrepare(t *testing.T) {
	i := domain.Installation{Provider: "steam", ExternalID: "42", Installed: true}
	spec, err := PrepareSteam(i)
	if err != nil || spec.URI != "steam://run/42" || spec.Confidence != "low" {
		t.Fatal(spec, err)
	}
	i.ExternalID = "42;calc.exe"
	if _, err := PrepareSteam(i); err == nil {
		t.Fatal("accepted invalid appid")
	}
	i.ExternalID = "42"
	i.Installed = false
	if _, err := PrepareSteam(i); err == nil {
		t.Fatal("accepted unavailable installation")
	}
}
