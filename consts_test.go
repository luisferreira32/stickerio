package stickerio

import "testing"

func Test_ConfigLoadBuildingMultipliers(t *testing.T) {
	testcases := []struct {
		name           string
		building       BuildingName
		level          int
		wantMultiplier float64
	}{
		{
			name:           "level zero multiplier",
			building:       Barracks,
			level:          0,
			wantMultiplier: 1.0,
		},
		{
			name:           "level three multiplier",
			building:       Barracks,
			level:          3,
			wantMultiplier: 0.75,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			if Config.Buildings[testcase.building].Multiplier[testcase.level] != testcase.wantMultiplier {
				t.Errorf(
					"unexpected multiplier for %s at level %d, wanted: %v, got: %v",
					testcase.building,
					testcase.level,
					testcase.wantMultiplier,
					Config.Buildings[testcase.building].Multiplier[testcase.level],
				)
			}
		})
	}
}
