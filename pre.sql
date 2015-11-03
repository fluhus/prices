-- Set cache size to 2 GB.
PRAGMA page_size = 4096;
PRAGMA default_cache_size = 524288;


----- TABLES -------------------------------------------------------------------

CREATE TABLE documentation (
-- Data dates from 8/10 - 20/10.
--
-- Changes from last version:
-- 1. Hopefully managed to translate gibberish text.
-- 2. Improved documentation of table fields.
--
-- See XML specifications here:
-- https://drive.google.com/file/d/0Bw2XXw9aHzlCT0xRMV9WSnZIS0E/view
--
-- See unresolved data issues here:
-- https://docs.google.com/document/d/168h95u3p-Xx7hSAHj_qKhM2aIo2G3vkmI3qDB1z-bNI/
--
-- PLEASE BE WARNED:
-- Vendors report garbage. Unless stated otherwise, any piece of information
-- presented in this database is brought 'as is' and should be treated as
-- unreliable, possibly incorrect, badly formatted, unsafe for use, offensive,
-- and inappropriate for children.
--
-- Data fields that were created or curated by us are marked explicitly as safe.

a
);

CREATE TABLE chains (
-- Maps chain code to chain name.
	chain_id   text PRIMARY KEY,   -- Chain code, as provided by GS1. (safe)
	chain_name text                -- Chain name in English. (safe)
);

INSERT INTO chains VALUES
('7290700100008','ColBo Hazi Hinam'),
('7290633800006','Coop'),
('7290492000005','Dor Alon'),
('7290055755557','Eden Teva Market'),
('7290876100000','Fresh Market'),
('7290785400000','Keshet Taamim'),
('7290661400001','Machsanei HaShuk'),
('7290058179503','Machsanei Lahav'),
('7290055700007','Mega'),
('7290103152017','Osher Ad'),
('7290058140886','Rami Levi Shivuk Shikma'),
('7290027600007','Shufersal'),
('7290873900009','Super Dosh'),
('7290803800003','Supershuk Yohananof'),
('7290873255550','Tiv Taam'),
('7290696200003','Victory'),
('7290725900003','Yeinot Bitan')
--('7290058140886','Zol VeBegadol')
;

-- TODO(amit): Add a bouncer for stores.
CREATE TABLE stores (
-- Identifies every store in the data. Each store may appear once.
	store_id           integer PRIMARY KEY AUTOINCREMENT,   -- (safe)
	chain_id           text  NOT NULL, -- Chain code, as provided by GS1.
	subchain_id        text  NOT NULL, -- Subchain number.
	reported_store_id  text  NOT NULL, -- Store number issued by the chain.
	CHECK  (chain_id <> '' AND subchain_id <> '' AND reported_store_id <> ''),
	UNIQUE (chain_id, subchain_id, reported_store_id)
);

CREATE TABLE stores_meta (
-- Metadata about stores. Each store may appear several times.
	timestamp        int,  -- Unix time when this entry was encountered.
	                       -- (safe)
	store_id         int   NOT NULL REFERENCES stores(store_id), -- (safe)
	bikoret_no       int,  -- ???
	store_type       int,  -- 1 for physical, 2 for online, 3 for both.
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
	item_id    integer PRIMARY KEY AUTOINCREMENT,  -- (safe)
	item_type  int   NOT NULL,  -- 0 for internal codes, 1 for barcodes.
	item_code  text  NOT NULL,  -- Barcode number or internal code.
	chain_id   text  NOT NULL,  -- Empty string for universal.
	CHECK  (item_code <> '' AND ((item_type = '0' AND chain_id <> '') OR
			(item_type = '1' AND chain_id = ''))),
	UNIQUE (item_type, item_code, chain_id)
);

CREATE TABLE items_meta (
-- Contains all metadata about each item. Each item may appear several times.
	timestamp                     int,  -- Unix time when this entry was
	                                    -- encountered. (safe)
	item_id                       int   NOT NULL REFERENCES items(item_id),
	                                    -- (safe)
	chain_id                      text  NOT NULL, -- Chain code, as provided by
	                                              -- GS1.
	update_time                   text,
	item_name                     text,
	manufacturer_item_description text,
	unit_quantity                 text,
	is_weighted                   text, -- 1 if sold in bulk, 0 if not.
	quantity_in_package           text, -- Quantity of units in a package.
	allow_discount                text, -- Is the item allowed in promotions.
	item_status                   text, -- ???
	crc                           int   -- Hash of all fields that need to be
	                                    -- compared for bouncing, to simplify
	                                    -- the trigger.
	                                    -- DO NOT USE FOR ANYTHING BUT THAT.
);

CREATE TABLE prices (
-- Contains all reported prices for all items.
	timestamp             int,   -- Unix time when this entry was encountered.
	                             -- (safe)
	item_id               int NOT NULL REFERENCES items(item_id),   -- (safe)
	store_id              int NOT NULL REFERENCES stores(store_id), -- (safe)
	price                 real,  -- Price in shekels as reported in raw data.
	unit_of_measure_price real,  -- Price in shekels as reported in raw data.
	unit_of_measure       text,  -- Gram, liter, etc.
	quantity              text,  -- How many grams/liters etc.
	crc                   int    -- Hash of all fields that need to be
	                             -- compared for bouncing, to simplify
	                             -- the trigger.
	                             -- DO NOT USE FOR ANYTHING BUT THAT.
);

-- TODO(amit): Check is_active field.
CREATE TABLE promos (
-- Identifies every promotion in the data. Promo id and metadata are saved
-- together since they are unique. A change in the metadata will be registered
-- as a new promo.
	promo_id                     integer PRIMARY KEY AUTOINCREMENT, -- (safe)
	timestamp_from               int,  -- Unix time when this entry was first
	                                   -- encountered. (safe)
	timestamp_to                 int,  -- Unix time when this entry was last
	                                   -- encountered + one day. (safe)
	chain_id                     text, -- Chain code, as provided by GS1.
	reward_type                  text, -- ???
	allow_multiple_discounts     text, -- 'Kefel mivtzaim'.
	promotion_id                 text, -- Issued by the chain, not by us.
	promotion_description        text,
	promotion_start_date         text,
	promotion_start_hour         text,
	promotion_end_date           text,
	promotion_end_hour           text,
	min_qty                      text, -- Min quantity for triggering promo.
	max_qty                      text, -- Max quantity for triggering promo.
	discount_rate                text,
	discount_type                text, -- 0 for relative, 1 for absolute.
	min_purchase_amnt            text,
	min_no_of_item_offered       text, -- Like min_qty, not sure what the
	                                   -- difference is.
	price_update_date            text,
	discounted_price             text,
	discounted_price_per_mida    text,
	additional_is_coupn          text, -- 1 if depends on coupon, 0 if not.
	additional_gift_count        text, -- Number of gift items.
	additional_is_total          text, -- Promo is on all items in the store.
	additional_min_basket_amount text, -- ???
	remarks                      text,
	number_of_items              int,  -- Number of items that take part in the
	                                   -- promotion. Should be equivalent to
	                                   -- count(*) on the id in promos_items,
	                                   -- but some of the promos are not
	                                   -- reported there. (safe)
	not_in_promos_items          int,  -- 0 if reported in promos_items, 1 if
	                                   -- not. (safe)
	crc                          int   -- Hash of all fields that need to be
	                                   -- compared for bouncing, to simplify
	                                   -- the trigger.
	                                   -- DO NOT USE FOR ANYTHING BUT THAT.
);

CREATE TABLE promos_stores (
-- Reports what stores take part in every promo. A single promo may have
-- several rows, one for each store.
	promo_id int NOT NULL REFERENCES promos(promo_id),  -- (safe)
	store_id int NOT NULL REFERENCES stores(store_id),  -- (safe)
	UNIQUE (promo_id, store_id)
);

CREATE TABLE promos_items (
-- Reports what items take part in every promo. A single promo may have
-- several rows, one for each item.
--
-- CAVEAT: promos that include more than 100 items are not reported here,
-- because those promos usually apply on an entire store ("everything for 10%
-- discount") and that bloats the DB. They are reported on the other tables
-- as usual.
	promo_id     int NOT NULL REFERENCES promos(promo_id), -- (safe)
	item_id      int NOT NULL REFERENCES items(item_id),   -- (safe)
	is_gift_item text,
	UNIQUE (promo_id, item_id)
);


----- INDEXES & TRIGGERS -------------------------------------------------------

CREATE INDEX prices_index ON prices(item_id, store_id, timestamp);

CREATE TRIGGER prices_bouncer
-- Prevents redundant rows from being added to the price table.
BEFORE INSERT ON prices FOR EACH ROW
WHEN new.crc = (
	SELECT crc FROM prices prices2 WHERE
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
-- Prevents redundant rows from being added to the promo table.
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


