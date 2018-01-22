-- Fetches all prices for a given barcode.
SELECT
timestamp, items.item_code item_code, prices.store_id, prices.price
FROM items, prices
WHERE
items.item_code = '{{BARCODE}}' AND
items.item_id = prices.item_id
ORDER BY prices.timestamp DESC;

