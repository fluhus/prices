-- Set cache size to 1 GB.
PRAGMA page_size = 4096;
PRAGMA cache_size = 262144;

CREATE TABLE stores_id (
-- Identifies every store in the data.
	id           integer PRIMARY KEY AUTOINCREMENT,
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
	id         integer PRIMARY KEY AUTOINCREMENT,
	item_type  int   NOT NULL,  -- 0 for internal barcodes, 1 for universal.
	item_code  text  NOT NULL,
	chain_id   text  NOT NULL,  -- Empty string for universal.
	CHECK  (item_code <> '' AND ((item_type = '0' AND chain_id <> '') OR
			(item_type = '1' AND chain_id = ''))),
	UNIQUE (item_type, item_code, chain_id)
);

CREATE TABLE items_meta (
-- Contains all metadata about each item.
	timestamp int,   -- Unix time (seconds since 1/1/1970).
	item_id                       int NOT NULL REFERENCES items_id(id),
	chain_id                      text NOT NULL,
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
	item_status                   text
);

CREATE TABLE prices (
-- Contains all reported prices for all items.
	timestamp int,   -- Unix time (seconds since 1/1/1970).
	item_id   int NOT NULL REFERENCES items_id(id),
	store_id  int NOT NULL REFERENCES stores_id(id),
	price     real,  -- Price in shekels as reported in raw data.
	CHECK (price >= 0)
);

CREATE INDEX prices_index ON prices(item_id, store_id, timestamp);

CREATE TRIGGER prices_bouncer
-- Prevents redundant rows from being added to the price table.
BEFORE INSERT ON prices FOR EACH ROW
WHEN new.price = (
	SELECT price FROM prices prices2 WHERE
	prices2.item_id = new.item_id AND
	prices2.store_id = new.store_id AND
	prices2.timestamp <= new.timestamp
	ORDER BY prices2.timestamp DESC LIMIT 1
)
BEGIN
	SELECT raise(ignore);
END;

CREATE INDEX items_meta_index ON items_meta(item_id, chain_id, timestamp);

CREATE TRIGGER items_bouncer
-- Prevents redundant rows from being added to the item table.
BEFORE INSERT ON items_meta FOR EACH ROW
WHEN
	new.item_name
	|| new.manufacturer_item_description
	|| new.unit_quantity
	|| new.quantity
	|| new.unit_of_measure
	|| new.is_weighted
	|| new.quantity_in_package
	|| new.allow_discount
	|| new.item_status = (
SELECT item_name
	|| manufacturer_item_description
	|| unit_quantity
	|| quantity
	|| unit_of_measure
	|| is_weighted
	|| quantity_in_package
	|| allow_discount
	|| item_status
	FROM items_meta items_meta2	WHERE
	items_meta2.item_id = new.item_id AND
	items_meta2.chain_id = new.chain_id AND
	items_meta2.timestamp <= new.timestamp
	ORDER BY items_meta2.timestamp DESC LIMIT 1
)
BEGIN
	SELECT raise(ignore);
END;



