.print
.print prices

select date(timestamp, 'unixepoch'), count(*) from prices
group by date(timestamp, 'unixepoch');

.print
.print items_meta

select date(timestamp, 'unixepoch'), count(*) from items_meta
group by date(timestamp, 'unixepoch');

.print
.print promos

select date(timestamp_from, 'unixepoch'), count(*) from promos
group by date(timestamp_from, 'unixepoch');

.print
.print promos_items

select promos.id, date(promos.timestamp_from, 'unixepoch'),
date(promos.timestamp_to, 'unixepoch'), count(*) from promos, promos_items
where promos.id = promos_items.promo_id
group by promos.id;

/*
.print
.print promos_stores

select promos.id, count(*) from promos, promos_stores
where promos.id = promos_stores.promo_id
group by promos.id;
*/

.print

select 'prices:       ' || count(*) from prices;
select 'items_id:     ' || count(*) from items_id;
select 'items_meta:   ' || count(*) from items_meta;
select 'promos:       ' || count(*) from promos;
select 'promos_items: ' || count(*) from promos_items;

.print

delete from items_meta where crc = (
select crc from items_meta items_meta2 where
items_meta2.item_id = items_meta.item_id and
items_meta2.chain_id = items_meta.chain_id and
items_meta2.timestamp <= items_meta.timestamp and
items_meta2.rowid <> items_meta.rowid
order by items_meta2.timestamp desc limit 1
);

delete from prices where price || unit_of_measure_price = (
select price || unit_of_measure_price from prices prices2 where
prices2.item_id = prices.item_id and
prices2.store_id = prices.store_id and
prices2.timestamp <= prices.timestamp and
prices2.rowid <> prices.rowid
order by prices2.timestamp desc limit 1
);

vacuum;

select 'prices:      ' || count(*) from prices;
select 'items_id:    ' || count(*) from items_id;
select 'items_meta:  ' || count(*) from items_meta;

/*
.print
.print "counts from 2015-07-01"
.headers on
select item_id, count(*) from items_meta
where date(timestamp, 'unixepoch') = '2015-07-01'
group by item_id having count(*) > 1;
--*/

/*
select * from items_meta
where item_id = 2 and date(timestamp, 'unixepoch') = '2015-07-01';
--*/


/*
CREATE TABLE prices_now AS
SELECT * FROM prices P WHERE timestamp = (
	SELECT max(timestamp) FROM prices R
	WHERE R.item_id = P.item_id AND R.store_id = P.store_id
);
*/

