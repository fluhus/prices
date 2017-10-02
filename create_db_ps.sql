-- Creates the database from scratch, from the parsed files.

\connect postgres

COMMENT ON DATABASE prices IS
'See XML specifications here:
https://drive.google.com/file/d/0Bw2XXw9aHzlCT0xRMV9WSnZIS0E/view

See unresolved data issues here:
https://docs.google.com/document/d/168h95u3p-Xx7hSAHj_qKhM2aIo2G3vkmI3qDB1z-bNI/

PLEASE BE WARNED:
Vendors report garbage. Unless stated otherwise, any piece of information presented in this database is brought ''as is'' and should be treated as unreliable, possibly incorrect, badly formatted, unsafe for use, offensive, and inappropriate for children.

Data fields that were created or curated by us are marked explicitly as safe.'
;

\connect prices

-- Create tables.
CREATE TABLE IF NOT EXISTS chains (
  chain_id   text PRIMARY KEY,
  chain_name text
);
DELETE FROM chains;
COMMENT ON TABLE chains IS 'Maps chain code to chain name.';
COMMENT ON COLUMN chains.chain_id IS 'Chain code, as provided by GS1. (safe)';
COMMENT ON COLUMN chains.chain_name IS 'Chain name in English. (safe)';

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

CREATE TABLE IF NOT EXISTS stores (
  store_id           integer,
  chain_id           text  NOT NULL,
  subchain_id        text  NOT NULL,
  reported_store_id  text  NOT NULL
);
DELETE FROM stores;
COMMENT ON TABLE stores IS 'Identifies every store in the data. Each store may appear once.';
COMMENT ON COLUMN stores.store_id IS '(safe)';
COMMENT ON COLUMN stores.chain_id IS 'Chain code, as provided by GS1.';
COMMENT ON COLUMN stores.subchain_id IS 'Subchain number.';
COMMENT ON COLUMN stores.reported_store_id IS 'Store number issued by the chain.';

COPY stores FROM '/home/amit/prices/data_parsed/stores.txt' WITH (FORMAT csv);

CREATE TABLE IF NOT EXISTS stores_meta (
  timestamp        int,
  store_id         int NOT NULL,
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
DELETE FROM stores_meta;
COMMENT ON TABLE stores_meta IS 'Metadata about stores. Each store may appear several times.';
COMMENT ON COLUMN stores_meta.timestamp IS 'Unix time when this entry was encountered. (safe)';
COMMENT ON COLUMN stores_meta.store_id IS 'References stores.store_id. (safe)';
COMMENT ON COLUMN stores_meta.bikoret_no IS '???';
COMMENT ON COLUMN stores_meta.store_type IS '1 for physical, 2 for online, 3 for both.';

COPY stores_meta FROM '/home/amit/prices/data_parsed/stores_meta.txt' WITH (FORMAT csv);

CREATE TABLE IF NOT EXISTS items (
  item_id    integer,
  item_type  int   NOT NULL,
  item_code  text  NOT NULL,
  chain_id   text
);
DELETE FROM items;
COMMENT ON TABLE items IS 'Identifies every commodity item in the data. Each item may appear once.';
COMMENT ON COLUMN items.item_id IS '(safe)';
COMMENT ON COLUMN items.item_type IS '0 for internal codes, 1 for barcodes.';
COMMENT ON COLUMN items.item_code IS 'Barcode number or internal code.';
COMMENT ON COLUMN items.chain_id IS 'Empty string for universal.';

COPY items FROM '/home/amit/prices/data_parsed/items.txt' WITH (FORMAT csv);

CREATE TABLE IF NOT EXISTS items_meta (
  timestamp                     int,
  item_id                       int  NOT NULL,
  chain_id                      text NOT NULL,
  update_time                   text,
  item_name                     text,
  manufacturer_item_description text,
  unit_quantity                 text,
  is_weighted                   text,
  quantity_in_package           text,
  allow_discount                text,
  item_status                   text
);
DELETE FROM items_meta;
COMMENT ON TABLE items_meta IS 'Contains all metadata about each item. Each item may appear several times.';
COMMENT ON COLUMN items_meta.timestamp IS 'Unix time when this entry was encountered. (safe)';
COMMENT ON COLUMN items_meta.item_id IS 'References items.item_id. (safe)';
COMMENT ON COLUMN items_meta.chain_id IS 'Chain code, as provided by GS1';
COMMENT ON COLUMN items_meta.is_weighted IS '1 if sold in bulk, 0 if not.';
COMMENT ON COLUMN items_meta.quantity_in_package IS 'Quantity of units in a package.';
COMMENT ON COLUMN items_meta.allow_discount IS 'Is the item allowed in promotions.';
COMMENT ON COLUMN items_meta.item_status IS '???';

COPY items_meta FROM '/home/amit/prices/data_parsed/items_meta.txt' WITH (FORMAT csv);

CREATE TABLE IF NOT EXISTS prices (
  timestamp             int,
  item_id               int NOT NULL,
  store_id              int NOT NULL,
  price                 real,
  unit_of_measure_price real,
  unit_of_measure       text,
  quantity              text
);
DELETE FROM prices;
COMMENT ON TABLE prices IS 'Contains all reported prices for all items.';
COMMENT ON COLUMN prices.timestamp IS 'Unix time when this entry was encountered. (safe)';
COMMENT ON COLUMN prices.item_id IS 'References items.item_id. (safe)';
COMMENT ON COLUMN prices.store_id IS 'References stores.store_id. (safe)';
COMMENT ON COLUMN prices.price IS 'Price in shekels as reported in raw data.';
COMMENT ON COLUMN prices.unit_of_measure_price IS 'Price in shekels as reported in raw data.';
COMMENT ON COLUMN prices.unit_of_measure IS 'Gram, liter, etc.';
COMMENT ON COLUMN prices.quantity IS 'How many grams/liters etc.';

COPY prices FROM '/home/amit/prices/data_parsed/prices.txt' WITH (FORMAT csv);

CREATE TABLE IF NOT EXISTS promos (
  promo_id                     integer,
  timestamp_from               int,
  timestamp_to                 int,
  chain_id                     text,
  promotion_id                 text,
  promotion_description        text,
  promotion_start_date         text,
  promotion_start_hour         text,
  promotion_end_date           text,
  promotion_end_hour           text,
  reward_type                  text,
  allow_multiple_discounts     text,
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
  number_of_items              int,
  not_in_promos_items          int
);
DELETE FROM promos;
COMMENT ON TABLE promos IS 'Identifies every promotion in the data. Promo id and metadata are saved together since they are unique. A change in the metadata will be registered as a new promo.';
COMMENT ON COLUMN promos.promo_id IS '(safe)';
COMMENT ON COLUMN promos.timestamp_from IS 'Unix time when this entry was first encountered. (safe)';
COMMENT ON COLUMN promos.timestamp_to IS 'Unix time when this entry was last encountered + one day. (safe)';
COMMENT ON COLUMN promos.chain_id IS 'Chain code, as provided by GS1.';
COMMENT ON COLUMN promos.promotion_id IS 'Issued by the chain, not by us.';
COMMENT ON COLUMN promos.reward_type IS '???';
COMMENT ON COLUMN promos.allow_multiple_discounts IS '''Kefel mivtzaim''.';
COMMENT ON COLUMN promos.min_qty IS 'Min quantity for triggering promo.';
COMMENT ON COLUMN promos.max_qty IS 'Max quantity for triggering promo.';
COMMENT ON COLUMN promos.discount_type IS '0 for relative, 1 for absolute.';
COMMENT ON COLUMN promos.min_no_of_item_offered IS 'Like min_qty, not sure what the difference is.';
COMMENT ON COLUMN promos.additional_is_coupn IS '1 if depends on coupon, 0 if not.';
COMMENT ON COLUMN promos.additional_gift_count IS 'Number of gift items.';
COMMENT ON COLUMN promos.additional_is_total IS 'Promo is on all items in the store.';
COMMENT ON COLUMN promos.additional_min_basket_amount IS '???';
COMMENT ON COLUMN promos.number_of_items IS 'Number of items that take part in the promotion. Should be equivalent to count(*) on the promo_id in promos_items, but some of the promos are not reported there. (safe)';
COMMENT ON COLUMN promos.not_in_promos_items IS '0 if reported in promos_items, 1 if not. (safe)';

COPY promos FROM '/home/amit/prices/data_parsed/promos.txt' WITH (FORMAT csv);

CREATE TABLE IF NOT EXISTS promos_stores (
  promo_id int NOT NULL,
  store_id int NOT NULL
);
DELETE FROM promos_stores;
COMMENT ON TABLE promos_stores IS 'Reports what stores take part in every promo. A single promo may have several rows, one for each store.';
COMMENT ON COLUMN promos_stores.promo_id IS 'References promos.promo_id. (safe)';
COMMENT ON COLUMN promos_stores.store_id IS 'References stores.store_id. (safe)';

COPY promos_stores FROM '/home/amit/prices/data_parsed/promos_stores.txt' WITH (FORMAT csv);

CREATE TABLE IF NOT EXISTS promos_items (
  promo_id     int NOT NULL,
  item_id      int NOT NULL,
  is_gift_item text
);
DELETE FROM promos_items;
COMMENT ON TABLE promos_items IS 'Reports what items take part in every promo. A single promo may have several rows, one for each item.

CAVEAT: promos that include more than 100 items are not reported here, because those promos usually apply on an entire store ("everything for 10% discount") and that bloats the DB. They are reported on the other tables as usual.';
COMMENT ON COLUMN promos_items.promo_id IS 'References promos.promo_id. (safe)';
COMMENT ON COLUMN promos_items.item_id IS 'References items.item_id. (safe)';

COPY promos_items FROM '/home/amit/prices/data_parsed/promos_items.txt' WITH (FORMAT csv);

CREATE TEMP TABLE promos_to (
-- A temporary table for updating the timestamp_to field in promos. The
-- timestamp_to field is evaluated after reporting each promo, so it has to be
-- updated separately.
  promo_id     int,
  timestamp_to int
);
COPY promos_to FROM '/home/amit/prices/data_parsed/promos_to.txt' WITH (FORMAT csv);
CREATE INDEX promos_to_index ON promos_to(promo_id, timestamp_to);

UPDATE promos SET timestamp_to = (
  SELECT max(timestamp_to) FROM promos_to
  WHERE promos_to.promo_id = promos.promo_id
);

