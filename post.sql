--CREATE INDEX prices_index ON prices(item_id, store_id, timestamp);

select 'prices: ' || count(*) from prices;
select 'items:  ' || count(*) from items_meta;

/*
CREATE TABLE prices_now AS
SELECT * FROM prices P WHERE timestamp = (
	SELECT max(timestamp) FROM prices R
	WHERE R.item_id = P.item_id AND R.store_id = P.store_id
);
*/

