package internal

// Inserted when a player starts a movement.
// Its processing generates an arrivalMovementEvent at a later epcoh (calculated based on distance between cities / speed).
type startMovementEvent struct {
	MovementID     string  `json:"movementID"`
	PlayerID       string  `json:"playerID"`
	OriginID       string  `json:"originID"`
	DestinationID  string  `json:"destinationID"`
	DepartureEpoch int64   `json:"departureEpoch"`
	Speed          float64 `json:"speed"`
	StickmenCount  int     `json:"stickmenCount"`
	SwordsmenCount int     `json:"swordsmenCount"`
	SticksCount    int     `json:"sticksCount"`
	CirclesCount   int     `json:"circlesCount"`
}

// Inserted when a startMovementEvent is processed.
// The arrival processing does the options:
// * reinforce if it's from the same player;
// * battle if it's from separate player
// * forage if it's abandoned
// Then, based on survival (if any), if the current city is not owned by the player (i.e., it was not a reinforce), re-calculate a startMovement event and schedule it.
type arrivalMovementEvent struct {
	MovementID     string `json:"movementID"`
	PlayerID       string `json:"playerID"`
	OriginID       string `json:"originID"`
	DestinationID  string `json:"destinationID"`
	SwordsmenCount int    `json:"swordsmenCount"`
	StickmenCount  int    `json:"stickmenCount"`
	SticksCount    int    `json:"sticksCount"`
	CirclesCount   int    `json:"circlesCount"`
	CityID         string `json:"cityID"`
}
