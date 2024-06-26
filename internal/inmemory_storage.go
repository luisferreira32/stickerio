package internal

type inMemoryStorage struct {
	cityList              map[tCityID]*city
	cityByCoordinates     map[coordinates]*city
	movementList          map[tMovementID]*movement
	unitQueuesPerCity     map[tCityID]map[tUnitQueueItemID]*unitQueueItem
	buildingQueuesPerCity map[tCityID]map[tBuildingQueueItemID]*buildingQueueItem
}

type coordinates struct {
	x tCoordinate
	y tCoordinate
}

func (m *inMemoryStorage) clear() {
	m.cityList = make(map[tCityID]*city)
	m.cityByCoordinates = make(map[coordinates]*city)
	m.movementList = make(map[tMovementID]*movement)
	m.unitQueuesPerCity = make(map[tCityID]map[tUnitQueueItemID]*unitQueueItem)
	m.buildingQueuesPerCity = make(map[tCityID]map[tBuildingQueueItemID]*buildingQueueItem)
}

func (m *inMemoryStorage) getCityByLocation(x, y tCoordinate) *city {
	return m.cityByCoordinates[coordinates{x: x, y: y}]
}

func (m *inMemoryStorage) createCity(cityID tCityID, c *city) {
	m.cityList[cityID] = c
	m.cityByCoordinates[coordinates{x: c.locationX, y: c.locationY}] = c
}

func (m *inMemoryStorage) deleteCity(cityID tCityID) {
	c := m.cityList[cityID]
	delete(m.cityByCoordinates, coordinates{x: c.locationX, y: c.locationY})
	delete(m.cityList, cityID)
	delete(m.buildingQueuesPerCity, cityID)
	delete(m.unitQueuesPerCity, cityID)
}
