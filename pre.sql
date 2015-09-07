-- Set cache size to 2 GB.
PRAGMA page_size = 4096;
PRAGMA default_cache_size = 524288;


----- TABLES -------------------------------------------------------------------

-- TODO(amit): Add a bouncer for stores.
CREATE TABLE stores (
-- Identifies every store in the data. Each store may appear once.
	id           integer PRIMARY KEY AUTOINCREMENT,
	chain_id     text  NOT NULL,
	subchain_id  text  NOT NULL,
	store_id     text  NOT NULL,
	CHECK  (chain_id <> '' AND subchain_id <> '' AND store_id <> ''),
	UNIQUE (chain_id, subchain_id, store_id)
);

CREATE TABLE stores_meta (
-- Metadata about stores. Each store may appear several times.
	timestamp        int, -- Unix time (seconds since 1/1/1970) when this entry
	                      -- was encountered.
	id               int  NOT NULL REFERENCES stores(id),
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

CREATE TABLE items (
-- Identifies every commodity item in the data. Each item may appear once.
	id         integer PRIMARY KEY AUTOINCREMENT,
	item_type  int   NOT NULL,  -- 0 for internal barcodes, 1 for universal.
	item_code  text  NOT NULL,
	chain_id   text  NOT NULL,  -- Empty string for universal.
	CHECK  (item_code <> '' AND ((item_type = '0' AND chain_id <> '') OR
			(item_type = '1' AND chain_id = ''))),
	UNIQUE (item_type, item_code, chain_id)
);

CREATE TABLE items_meta (
-- Contains all metadata about each item. Each item may appear several times.
	timestamp                     int, -- Unix time (seconds since 1/1/1970) 
	                                   -- when this entry was encountered.
	item_id                       int  NOT NULL REFERENCES items(id),
	chain_id                      text NOT NULL,
	update_time                   text,
	item_name                     text,
	manufacturer_item_description text,
	unit_quantity                 text,
	quantity                      text,
	unit_of_measure               text,
	is_weighted                   text,
	quantity_in_package           text,
	allow_discount                text,
	item_status                   text,
	crc                           int  -- Hash of all fields that need to be
	                                   -- compared for bouncing, to simplify
	                                   -- the trigger.
	                                   -- DO NOT USE FOR ANYTHING BUT THAT.
);

CREATE TABLE prices (
-- Contains all reported prices for all items.
	timestamp             int,   -- Unix time (seconds since 1/1/1970) when this
	                             -- entry was encountered.
	item_id               int NOT NULL REFERENCES items(id),
	store_id              int NOT NULL REFERENCES stores(id),
	price                 real,  -- Price in shekels as reported in raw data.
	unit_of_measure_price real,  -- Price in shekels as reported in raw data.
	CHECK (price >= 0 AND unit_of_measure_price >= 0)
);

-- TODO(amit): Check is_active field.
CREATE TABLE promos (
-- Identifies every promotion in the data. Promo id and metadata are saved
-- together since they are unique. A change in the metadata will be registered
-- as a new promo.
--
-- All timestamps are unix time (seconds since 1/1/1970).
	id integer PRIMARY KEY AUTOINCREMENT,
	timestamp_from               int,  -- Time when this entry was first
	                                   -- encountered.
	timestamp_to                 int,  -- Time when this entry was first
	                                   -- encountered + one day.
	chain_id                     text,
	reward_type                  text,
	allow_multiple_discounts     text,
	promotion_id                 text,
	promotion_description        text,
	promotion_start_date         text,
	promotion_start_hour         text,
	promotion_end_date           text,
	promotion_end_hour           text,
	min_qty                      text,
	max_qty                      text,
	discount_rate                text,
	discount_type                text,
	min_purchase_amnt            text,
	min_no_of_item_offered       text,
	price_update_date            text,
	discounted_price             text,
	discounted_price_per_mida    text,
	additional_is_coupn          text,
	additional_gift_count        text,
	additional_is_total          text,
	additional_min_basket_amount text,
	remarks                      text,
	crc                          int   -- Hash of all fields that need to be
	                                   -- compared for bouncing, to simplify
	                                   -- the trigger.
	                                   -- DO NOT USE FOR ANYTHING BUT THAT.
);

CREATE TABLE promos_stores (
-- Reports what stores take part in every promo. A single promo may have
-- several rows, one for each store.
	promo_id int NOT NULL REFERENCES promos(id),
	store_id int NOT NULL REFERENCES stores(id),
	UNIQUE (promo_id, store_id)
);

CREATE TABLE promos_items (
-- Reports what items take part in every promo. A single promo may have
-- several rows, one for each item.
	promo_id     int NOT NULL REFERENCES promos(id),
	item_id      int NOT NULL REFERENCES items(id),
	is_gift_item text,
	UNIQUE (promo_id, item_id)
);


----- INDEXES & TRIGGERS -------------------------------------------------------

CREATE INDEX prices_index ON prices(item_id, store_id, timestamp);

CREATE TRIGGER prices_bouncer
-- Prevents redundant rows from being added to the price table.
BEFORE INSERT ON prices FOR EACH ROW
WHEN new.price || new.unit_of_measure_price = (
	SELECT price || unit_of_measure_price FROM prices prices2 WHERE
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
WHEN new.crc = (
	SELECT crc
	FROM items_meta items_meta2 WHERE
	items_meta2.item_id = new.item_id AND
	items_meta2.chain_id = new.chain_id AND
	items_meta2.timestamp <= new.timestamp
	ORDER BY items_meta2.timestamp DESC LIMIT 1
)
BEGIN
	SELECT raise(ignore);
END;

CREATE INDEX promos_index ON promos(crc, chain_id, promotion_id);

CREATE TRIGGER promos_bouncer
-- Prevents redundant rows from being added to the item table.
BEFORE INSERT ON promos FOR EACH ROW
WHEN (
	SELECT rowid FROM promos promos2 WHERE
	promos2.crc = new.crc AND
	promos2.chain_id = new.chain_id AND
	promos2.promotion_id = new.promotion_id
) IS NOT NULL
BEGIN
	UPDATE promos SET
		timestamp_to = max(timestamp_to, new.timestamp_to),
		timestamp_from = min(timestamp_from, new.timestamp_from)
	WHERE
		crc = new.crc AND
		chain_id = new.chain_id AND
		promotion_id = new.promotion_id;
	SELECT raise(ignore);
END;


