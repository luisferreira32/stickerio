package internal

type inMemoryStorage struct {
	cityList              map[tCityID]*city
	movementList          map[tMovementID]*movement
	unitQueuesPerCity     map[tCityID]map[tItemID]*unitQueueItem
	buildingQueuesPerCity map[tCityID]map[tItemID]*buildingQueueItem
}

func (m *inMemoryStorage) clear() {
	m.cityList = make(map[tCityID]*city)
	m.movementList = make(map[tMovementID]*movement)
	m.unitQueuesPerCity = make(map[tCityID]map[tItemID]*unitQueueItem)
	m.buildingQueuesPerCity = make(map[tCityID]map[tItemID]*buildingQueueItem)
}
