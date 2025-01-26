package gounix_test

import (
	"testing"

	"github.com/mekramy/gounix"
)

func TestCronGenerator(t *testing.T) {
	data := map[string]gounix.CronJob{
		"@reboot do some": gounix.NewCronJob("do some", nil).AtReboot(),
		"30 20 * * 0 do some": gounix.
			NewCronJob("do some", gounix.NewTZ().Hour(3).Minute(30)).
			Weekly(gounix.Auto),
		"0 03 * * 3 do some": gounix.
			NewCronJob("do some", gounix.NewTZ().Hour(0).Minute(0).Weekend(gounix.Wednesday)).
			Weekly(gounix.Auto).
			SetHour(3).
			SetMinute(0),
	}

	for expected, cron := range data {
		result := cron.Compile()
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		} else {
			t.Logf("Test passed on %s", expected)
		}
	}
}
