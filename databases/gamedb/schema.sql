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
    origin_id text,
    destination_id text,

    r_circles_count int,
    r_stick_count int,
    u_stickmen_count int,
    u_swordmen_count int
);