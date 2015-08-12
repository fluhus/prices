select 'prices:      ' || count(*) from prices;
select 'items_id:    ' || count(*) from items_id;
select 'items_meta:  ' || count(*) from items_meta;

.print
.print prices

select date(timestamp, 'unixepoch'), count(*) from prices
group by date(timestamp, 'unixepoch');

.print
.print items_meta

select date(timestamp, 'unixepoch'), count(*) from items_meta
group by date(timestamp, 'unixepoch');

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

