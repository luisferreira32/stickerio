package stickerio

type StickerioPlayer struct {
	ID       string
	Nickname string
	Score    int
}

type UnitCount struct {
	StickmenCount  int
	SwordsmenCount int
}

type ListCityInfoFilters struct {
	LastCityID string
	PageSize   int
	PlayerID   string
	// List cities in the area of (x1, y1, x2, y2)
	LocationBounds []int
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
	*UnitCount
}

type ResourcesCount struct {
	SticksCount  int
	CirclesCount int
}

type ListMovementsFilters struct {
	LastMovementID string
	PageSize       int
	OriginCityID   string
}

type Movement struct {
	ID             string
	PlayerID       string
	OriginID       string
	DestinationID  string
	DepartureEpoch int64
	Speed          float64
	*UnitCount
	*ResourcesCount
}

type ListUnitQueueItemFilters struct {
	LastMovementID string
	PageSize       int
}

type UnitQueueItem struct {
	ID          string
	CityID      string
	QueuedEpoch int64
	DurationSec int
	UnitCount   int
	UnitType    UnitName
}

type LisBuildingQueueItemFilters struct {
	LastMovementID string
	PageSize       int
}

type BuildingQueueItem struct {
	ID             string
	CityID         string
	QueuedEpoch    int64
	DurationSec    int
	TargetLevel    int
	TargetBuilding BuildingName
}
