package stickerio

var (
	MineResourceCosts = map[int]*Resources{
		1: {SticksCount: 1},
		2: {SticksCount: 10},
		3: {SticksCount: 100},
		4: {SticksCount: 200},
		5: {SticksCount: 1000},
	}
	MineResourceProductionPerSecond = map[int]*Resources{
		1: {SticksCount: 1},
		2: {SticksCount: 2},
		3: {SticksCount: 4},
		4: {SticksCount: 8},
		5: {SticksCount: 16},
	}
)

const (
	SwordsmenSpeed = 1.0
)
