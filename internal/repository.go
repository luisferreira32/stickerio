package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

func NewStickerioRepository(dataSourceName string) *StickerioRepository {
	db, err := sql.Open("sqlite", dataSourceName) // TODO: don't use sqlite
	if err != nil {
		log.Fatal(err)
	}
	return &StickerioRepository{
		db: db,
	}
}

type StickerioRepository struct {
	db *sql.DB
}

type event struct {
	id      string
	name    string
	epoch   int64
	payload string // TODO: json encode might not be the most efficient way - improve this
}

func (r *StickerioRepository) InsertEvent(ctx context.Context, e *event) error {
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

func (r *StickerioRepository) ListEvents(ctx context.Context, untilEpoch int64) ([]*event, error) {
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

type dbCity struct {
	id            string
	name          string
	playerID      string
	locationX     int32
	locationY     int32
	mineLevel     int32
	barracksLevel int32
	resourceBase  string
	resourceEpoch int64
	unitCount     string
}

func (r *StickerioRepository) GetCity(ctx context.Context, id, playerID string) (*dbCity, error) {
	const getCityQuery = `
SELECT
id,
city_name,
player_id,
location_x,
location_y,
b_mine_level,
b_barracks_level,
r_base,
r_epoch,
u_count
FROM cities_view
WHERE id=$1 AND player_id=$2
`

	rows, err := r.db.QueryContext(ctx, getCityQuery, id, playerID)
	if err != nil {
		return nil, fmt.Errorf("getCityQuery failed: %w", err)
	}

	result := &dbCity{}

	for rows.Next() {
		err := rows.Scan(
			&result.id,
			&result.name,
			&result.playerID,
			&result.locationX,
			&result.locationY,
			&result.mineLevel,
			&result.barracksLevel,
			&result.resourceBase,
			&result.resourceEpoch,
			&result.unitCount,
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

func (r *StickerioRepository) GetCityInfo(ctx context.Context, id string) (*dbCity, error) {
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

	result := &dbCity{}

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

func (r *StickerioRepository) ListCityInfo(ctx context.Context, lastID string, pageSize int, filters ...listCityInfoFilterOpt) ([]*dbCity, error) {
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

	results := make([]*dbCity, 0, pageSize)

	for rows.Next() {
		result := &dbCity{}
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

type dbMovement struct {
	id             string
	playerID       string
	originID       string
	destinationID  string
	departureEpoch int64
	speed          float32
	resourceCount  string
	unitCount      string
}

func (r *StickerioRepository) GetMovement(ctx context.Context, id, playerID string) (*dbMovement, error) {
	const getMovementQuery = `
SELECT
id,
player_id,
origin_id,
destination_id,
departure_epoch,
speed,
r_count,
u_count
FROM movements_view
WHERE id=$1 AND player_id=$2
`

	rows, err := r.db.QueryContext(ctx, getMovementQuery, id, playerID)
	if err != nil {
		return nil, fmt.Errorf("getMovementQuery failed: %w", err)
	}

	result := &dbMovement{}

	for rows.Next() {
		err := rows.Scan(
			&result.id,
			&result.playerID,
			&result.originID,
			&result.destinationID,
			&result.departureEpoch,
			&result.speed,
			&result.resourceCount,
			&result.unitCount,
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

func (r *StickerioRepository) ListMovements(ctx context.Context, playerID, lastID string, pageSize int, filters ...listMovementsFilterOpt) ([]*dbMovement, error) {
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
r_count,
u_count
FROM movements_view
%s
ORDER BY id
LIMIT $3
`, filtersQuery)

	rows, err := r.db.QueryContext(ctx, listMovementQuery, filtersValues...)
	if err != nil {
		return nil, fmt.Errorf("listMovementQuery failed: %w", err)
	}

	results := make([]*dbMovement, 0, pageSize)

	for rows.Next() {
		result := &dbMovement{}
		err := rows.Scan(
			&result.id,
			&result.playerID,
			&result.originID,
			&result.destinationID,
			&result.departureEpoch,
			&result.speed,
			&result.resourceCount,
			&result.unitCount,
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

func (r *StickerioRepository) UpsertMovement(ctx context.Context, m *dbMovement) error {
	const upsertMovementQuery = `
INSERT INTO movements_view(
id,
player_id,
origin_id,
destination_id,
departure_epoch,
speed,
r_count,
u_count)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT(id) REPLACE
`

	_, err := r.db.ExecContext(
		ctx,
		upsertMovementQuery,
		m.id,
		m.playerID,
		m.originID,
		m.destinationID,
		m.departureEpoch,
		m.speed,
		m.resourceCount,
		m.unitCount,
	)
	if err != nil {
		return fmt.Errorf("upsertMovementQuery failed: %w", err)
	}

	return nil
}

type dbUnitQueueItem struct {
	id          string
	cityID      string
	queuedEpoch int64
	durationSec int32
	unitCount   int32
	unitType    string
}

func (r *StickerioRepository) GetUnitQueueItem(ctx context.Context, id, cityID string) (*dbUnitQueueItem, error) {
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

	result := &dbUnitQueueItem{}

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

func (r *StickerioRepository) ListUnitQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*dbUnitQueueItem, error) {
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

	results := make([]*dbUnitQueueItem, 0, pageSize)

	for rows.Next() {
		result := &dbUnitQueueItem{}
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

type dbBuildingQueueItem struct {
	id             string
	cityID         string
	queuedEpoch    int64
	durationSec    int32
	targetLevel    int32
	targetBuilding string
}

func (r *StickerioRepository) GetBuildingQueueItem(ctx context.Context, id, cityID string) (*dbBuildingQueueItem, error) {
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

	result := &dbBuildingQueueItem{}

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

func (r *StickerioRepository) ListBuildingQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*dbBuildingQueueItem, error) {
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

	results := make([]*dbBuildingQueueItem, 0, pageSize)

	for rows.Next() {
		result := &dbBuildingQueueItem{}
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
