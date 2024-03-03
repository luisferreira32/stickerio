package stickerio

type StickerioPlayer struct {
	ID       string
	Nickname string
	Score    int
}

type Units struct {
	StickmenCount  int
	SwordsmenCount int
}

type CityInfo struct {
	ID        string
	Name      string
	PlayerID  string
	LocationX int
	LocationY int
}

type CityBuildings struct {
	BarracksLevel int
	MineLevel     int
}

type CityResources struct {
	SticksCountBase   int
	SticksCountEpoch  int64
	CirclesCountBase  int
	CirclesCountEpoch int64
}

type City struct {
	*CityInfo
	*CityBuildings
	*CityResources
	*Units
}

type Resources struct {
	SticksCount  int
	CirclesCount int
}

type Movements struct {
	ID            string
	OriginID      string
	DestinationID string

	*Units
	*Resources
}
