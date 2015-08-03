CREATE TABLE stores_id (
-- Identifies every store in the data.
	id           integer PRIMARY KEY,
	chain_id     text  NOT NULL,
	subchain_id  text  NOT NULL,
	store_id     text  NOT NULL,
	CHECK  (chain_id <> '' AND subchain_id <> '' AND store_id <> ''),
	UNIQUE (chain_id, subchain_id, store_id)
);

CREATE TABLE stores_meta (
-- Metadata about stores.
	timestamp int,  -- Unix time (seconds since 1/1/1970).
	id               int NOT NULL REFERENCES stores_id(id),
	bikoret_no       int,
	store_type       int,
	chain_name       text,
	subchain_name    text,
	store_name       text,
	address          text,
	city             text,
	zip_code         text,
	last_update_date text,
	last_update_time text
);

CREATE TABLE items_id (
-- Identifies every commodity item in the data.
	id         integer PRIMARY KEY,
	item_type  int   NOT NULL,  -- 0 for internal barcodes, 1 for universal.
	item_code  text  NOT NULL,
	chain_id   text  NOT NULL,  -- Empty string for universal.
	CHECK  (item_code <> '' AND ((item_type = '0' AND chain_id <> '') OR
			(item_type = '1' AND chain_id = ''))),
	UNIQUE (item_type, item_code, chain_id)
);

CREATE TABLE items_meta (
-- Contains the latest report of each store on each item.
	timestamp int,   -- Unix time (seconds since 1/1/1970).
	item_id                       int NOT NULL REFERENCES items_id(id),
	store_id                      int NOT NULL REFERENCES stores_id(id),
	update_time                   text,
	item_name                     text,
	manufacturer_name             text,
	manufacturer_country          text,
	manufacturer_item_description text,
	unit_quantity                 text,
	quantity                      text,
	unit_of_measure               text,
	is_weighted                   text,
	quantity_in_package           text,
	unit_of_measure_price         text,
	allow_discount                text,
	item_status                   text,
	UNIQUE (item_id, store_id)
);

CREATE TABLE prices (
-- Contains all reported prices for all items.
	timestamp int,   -- Unix time (seconds since 1/1/1970).
	item_id   int NOT NULL REFERENCES items_id(id),
	store_id  int NOT NULL REFERENCES stores_id(id),
	price     real,  -- Price in shekels as reported in raw data.
	CHECK (price >= 0)
);

