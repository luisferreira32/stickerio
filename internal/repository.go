package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

func NewStickerioRepository(dataSourceName string) *StickerioRepositoryImpl {
	db, err := sql.Open("sqlite", dataSourceName) // TODO: don't use sqlite
	if err != nil {
		log.Fatal(err)
	}
	return &StickerioRepositoryImpl{
		db: db,
	}
}

type StickerioRepositoryImpl struct {
	db *sql.DB
}

type event struct {
	id      string
	name    string
	epoch   int64
	payload string // TODO: json encode might not be the most efficient way - improve this
}

func (r *StickerioRepositoryImpl) InsertEvent(ctx context.Context, e *event) error {
	const insertEventQuery = `
INSERT INTO event_source(id, event_name, epoch, payload) VALUES ($1, $2, $3, $4)
ON CONFLICT(id) DO NOTHING
`
	_, err := r.db.ExecContext(ctx, insertEventQuery, e.id, e.name, e.epoch, e.payload)
	if err != nil {
		return fmt.Errorf("insertEventQuery failed: %w", err)
	}
	return nil
}

func (r *StickerioRepositoryImpl) ListEvents(ctx context.Context, untilEpoch int64) ([]*event, error) {
	const listEventsQuery = `
SELECT
id,
event_name,
epoch,
payload
FROM event_source
WHERE epoch <= $1
ORDER BY epoch, id
`
	rows, err := r.db.QueryContext(ctx, listEventsQuery, untilEpoch)
	if err != nil {
		return nil, fmt.Errorf("listEventsQuery failed: %w", err)
	}

	results := make([]*event, 0)

	for rows.Next() {
		result := &event{}
		err := rows.Scan(
			&result.id,
			&result.name,
			&result.epoch,
			&result.payload,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

type city struct {
	id                string
	name              string
	playerID          string
	locationX         int32
	locationY         int32
	mineLevel         int32
	barracksLevel     int32
	sticksCountBase   int64
	sticksCountEpoch  int64
	circlesCountBase  int64
	circlesCountEpoch int64
	stickmenCount     int32
	swordsmenCount    int32
}

func (r *StickerioRepositoryImpl) GetCity(ctx context.Context, id, playerID string) (*city, error) {
	const getCityQuery = `
SELECT
id,
city_name,
player_id,
location_x,
location_y,
b_mine_level,
b_barracks_level,
r_stick_count_base,
r_stick_count_epoch,
r_circles_count_base,
r_circles_count_epoch,
u_stickmen_count,
u_swordmen_count
FROM cities_view
WHERE id=$1 AND player_id=$2
`

	rows, err := r.db.QueryContext(ctx, getCityQuery, id, playerID)
	if err != nil {
		return nil, fmt.Errorf("getCityQuery failed: %w", err)
	}

	result := &city{}

	for rows.Next() {
		err := rows.Scan(
			&result.id,
			&result.name,
			&result.playerID,
			&result.locationX,
			&result.locationY,
			&result.mineLevel,
			&result.barracksLevel,
			&result.sticksCountBase,
			&result.sticksCountEpoch,
			&result.circlesCountBase,
			&result.circlesCountEpoch,
			&result.stickmenCount,
			&result.swordsmenCount,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return result, nil
}

func (r *StickerioRepositoryImpl) GetCityInfo(ctx context.Context, id string) (*city, error) {
	const getCityInfoQuery = `
SELECT
id,
city_name,
player_id,
location_x,
location_y
FROM cities_view
WHERE id=$1
`

	rows, err := r.db.QueryContext(ctx, getCityInfoQuery, id)
	if err != nil {
		return nil, fmt.Errorf("getCityInfoQuery failed: %w", err)
	}

	result := &city{}

	for rows.Next() {
		err := rows.Scan(
			&result.id,
			&result.name,
			&result.playerID,
			&result.locationX,
			&result.locationY,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return result, nil
}

type listCityInfoFilterOpt func() (string, interface{})

func withPlayerID(playerID string) listCityInfoFilterOpt {
	return func() (string, interface{}) {
		return "player_id=", playerID
	}
}

func withinLocation(x1, y1, x2, y2 int) []listCityInfoFilterOpt {
	return []listCityInfoFilterOpt{
		func() (string, interface{}) {
			return "location_x>=", x1
		},
		func() (string, interface{}) {
			return "location_y>=", y1
		},
		func() (string, interface{}) {
			return "location_x<=", x2
		},
		func() (string, interface{}) {
			return "location_y<=", y2
		},
	}
}

func (r *StickerioRepositoryImpl) ListCityInfo(ctx context.Context, lastID string, pageSize int, filters ...listCityInfoFilterOpt) ([]*city, error) {
	filtersQuery := "WHERE id>$1"
	filtersValues := make([]interface{}, 0, len(filters))
	filtersValues = append(filtersValues, lastID, pageSize)
	for i, filter := range filters {
		q, v := filter()
		filtersValues = append(filtersValues, v)
		filtersQuery += " AND " + q + "$" + strconv.Itoa(i+3) // starts at $3 since $1 and $2 are already taken for lastID and pagination
	}

	listCityInfoQuery := fmt.Sprintf(`
SELECT
id,
city_name,
player_id,
location_x,
location_y
FROM cities_view
%s
ORDER BY id
LIMIT $2
`, filtersQuery)

	rows, err := r.db.QueryContext(ctx, listCityInfoQuery, filtersValues...)
	if err != nil {
		return nil, fmt.Errorf("listCityInfoQuery failed: %w", err)
	}

	results := make([]*city, 0, pageSize)

	for rows.Next() {
		result := &city{}
		err := rows.Scan(
			&result.id,
			&result.name,
			&result.playerID,
			&result.locationX,
			&result.locationY,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

type movement struct {
	id             string
	playerID       string
	originID       string
	destinationID  string
	departureEpoch int64
	speed          float32
	circlesCount   int32
	stickCount     int32
	stickmenCount  int32
	swordmenCount  int32
}

func (r *StickerioRepositoryImpl) GetMovement(ctx context.Context, id, playerID string) (*movement, error) {
	const getMovementQuery = `
SELECT
id,
player_id,
origin_id,
destination_id,
departure_epoch,
speed,
r_circles_count,
r_stick_count,
u_stickmen_count,
u_swordmen_count
FROM movements_view
WHERE id=$1 AND player_id=$2
`

	rows, err := r.db.QueryContext(ctx, getMovementQuery, id, playerID)
	if err != nil {
		return nil, fmt.Errorf("getMovementQuery failed: %w", err)
	}

	result := &movement{}

	for rows.Next() {
		err := rows.Scan(
			&result.id,
			&result.playerID,
			&result.originID,
			&result.destinationID,
			&result.departureEpoch,
			&result.speed,
			&result.circlesCount,
			&result.stickCount,
			&result.stickmenCount,
			&result.swordmenCount,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return result, nil
}

type listMovementsFilterOpt func() (string, interface{})

func withOriginCityID(cityID string) listMovementsFilterOpt {
	return func() (string, interface{}) {
		return "origin_id=", cityID
	}
}

func (r *StickerioRepositoryImpl) ListMovements(ctx context.Context, playerID, lastID string, pageSize int, filters ...listMovementsFilterOpt) ([]*movement, error) {
	filtersQuery := "WHERE player_id=$1 AND id>$1"
	filtersValues := make([]interface{}, 0, len(filters))
	filtersValues = append(filtersValues, playerID, lastID, pageSize)
	for i, filter := range filters {
		q, v := filter()
		filtersValues = append(filtersValues, v)
		filtersQuery += " AND " + q + "$" + strconv.Itoa(i+4) // starts at $4 since $1 to $3 are already taken for lastID, playerID and pagination
	}

	listMovementQuery := fmt.Sprintf(`
SELECT
id,
player_id,
origin_id,
destination_id,
departure_epoch,
speed,
r_circles_count,
r_stick_count,
u_stickmen_count,
u_swordmen_count
FROM movements_view
%s
ORDER BY id
LIMIT $3
`, filtersQuery)

	rows, err := r.db.QueryContext(ctx, listMovementQuery, filtersValues...)
	if err != nil {
		return nil, fmt.Errorf("listMovementQuery failed: %w", err)
	}

	results := make([]*movement, 0, pageSize)

	for rows.Next() {
		result := &movement{}
		err := rows.Scan(
			&result.id,
			&result.playerID,
			&result.originID,
			&result.destinationID,
			&result.departureEpoch,
			&result.speed,
			&result.circlesCount,
			&result.stickCount,
			&result.stickmenCount,
			&result.swordmenCount,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

type unitQueueItem struct {
	id          string
	cityID      string
	queuedEpoch int64
	durationSec int32
	unitCount   int32
	unitType    string
}

func (r *StickerioRepositoryImpl) GetUnitQueueItem(ctx context.Context, id, cityID string) (*unitQueueItem, error) {
	const getUnitQueueItemQuery = `
SELECT
id,
city_id,
queued_epoch,
duration_s,
unit_count,
unit_type
FROM unit_queue_view
WHERE id=$1 AND city_id=$2
`

	rows, err := r.db.QueryContext(ctx, getUnitQueueItemQuery, id, cityID)
	if err != nil {
		return nil, fmt.Errorf("getUnitQueueItemQuery failed: %w", err)
	}

	result := &unitQueueItem{}

	for rows.Next() {
		err := rows.Scan(
			&result.id,
			&result.cityID,
			&result.queuedEpoch,
			&result.durationSec,
			&result.unitCount,
			&result.unitType,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return result, nil
}

func (r *StickerioRepositoryImpl) ListUnitQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*unitQueueItem, error) {
	filtersValues := []interface{}{cityID, lastID, pageSize}
	const listUnitQueueItemsQuery = `
SELECT
id,
city_id,
queued_epoch,
duration_s,
unit_count,
unit_type
FROM unit_queue_view
WHERE city_id=$1 AND id>$2
ORDER BY id
LIMIT $3
`

	rows, err := r.db.QueryContext(ctx, listUnitQueueItemsQuery, filtersValues...)
	if err != nil {
		return nil, fmt.Errorf("listUnitQueueItemsQuery failed: %w", err)
	}

	results := make([]*unitQueueItem, 0, pageSize)

	for rows.Next() {
		result := &unitQueueItem{}
		err := rows.Scan(
			&result.id,
			&result.cityID,
			&result.queuedEpoch,
			&result.durationSec,
			&result.unitCount,
			&result.unitType,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

type buildingQueueItem struct {
	id             string
	cityID         string
	queuedEpoch    int64
	durationSec    int32
	targetLevel    int32
	targetBuilding string
}

func (r *StickerioRepositoryImpl) GetBuildingQueueItem(ctx context.Context, id, cityID string) (*buildingQueueItem, error) {
	const getUnitQueueItemQuery = `
SELECT
id,
city_id,
queued_epoch,
duration_s,
target_level,
target_building
FROM unit_queue_view
WHERE id=$1 AND city_id=$2
`

	rows, err := r.db.QueryContext(ctx, getUnitQueueItemQuery, id, cityID)
	if err != nil {
		return nil, fmt.Errorf("getUnitQueueItemQuery failed: %w", err)
	}

	result := &buildingQueueItem{}

	for rows.Next() {
		err := rows.Scan(
			&result.id,
			&result.cityID,
			&result.queuedEpoch,
			&result.durationSec,
			&result.targetLevel,
			&result.targetBuilding,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return result, nil
}

func (r *StickerioRepositoryImpl) ListBuildingQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*buildingQueueItem, error) {
	filtersValues := []interface{}{cityID, lastID, pageSize}
	const listBuildingQueueItemsQuery = `
SELECT
id,
city_id,
queued_epoch,
duration_s,
target_level,
target_building
FROM unit_queue_view
WHERE city_id=$1 AND id>$2
ORDER BY id
LIMIT $3
`

	rows, err := r.db.QueryContext(ctx, listBuildingQueueItemsQuery, filtersValues...)
	if err != nil {
		return nil, fmt.Errorf("listBuildingQueueItemsQuery failed: %w", err)
	}

	results := make([]*buildingQueueItem, 0, pageSize)

	for rows.Next() {
		result := &buildingQueueItem{}
		err := rows.Scan(
			&result.id,
			&result.cityID,
			&result.queuedEpoch,
			&result.durationSec,
			&result.targetLevel,
			&result.targetBuilding,
		)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}
