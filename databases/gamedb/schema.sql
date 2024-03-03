-----
-- sources of truth
-----

create table if not exists event_source (
    id text primary key,
    epoch int,
    payload text
);

create table if not exists players (
    id text primary key,
    nickname text,
    score int
);

-----
-- view tables
-----

create table if not exists cities_view (
    id text primary key,
    city_name text,
    player_id text,
    location_x int,
    location_y int,
    b_mine_level int,
    b_barracks_level int,
    r_stick_count_base int,
    r_stick_count_epoch int,
    r_circles_count_base int,
    r_circles_count_epoch int,
    u_stickmen_count int,
    u_swordmen_count int
);

create table if not exists movements_view (
    id text primary key,
    player_id text,
    origin_id text,
    destination_id text,
    departure_epoch int,
    speed real,
    r_circles_count int,
    r_stick_count int,
    u_stickmen_count int,
    u_swordmen_count int
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