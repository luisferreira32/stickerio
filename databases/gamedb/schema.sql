create table if not exists event_source (
    id text primary key,
    event_name text,
    epoch int,
    payload text -- json serialization of the events
);

create table if not exists players (
    id text primary key,
    nickname text,
    score int
);

create table if not exists cities_view (
    id text primary key,
    city_name text,
    player_id text,
    location_x int,
    location_y int,
    b_economic_level text, -- json serialization of economic buildingID: level
    b_military_level text, -- json serialization of military buildingID: level
    r_base text, -- json serialization of resourceID: baseQuantity
    r_epoch int,
    u_count text -- json serialization of unitID: count
);

create table if not exists movements_view (
    id text primary key,
    player_id text,
    origin_id text,
    destination_id text,
    destination_x int,
    destination_y int,
    departure_epoch int,
    speed real,
    r_count text, -- json serialization of resourceID: count
    u_count text -- json serialization of unitID: count
);

create table if not exists unit_queue_view (
    id text primary key,
    city_id text,
    queued_epoch int,
    duration_s int,
    unit_count int,
    unit_type text
);

create table if not exists building_queue_view (
    id text primary key,
    city_id text,
    queued_epoch int,
    duration_s int,
    target_level int,
    target_building text
);
