SELECT item_id, chain_id, item_name, count(*) count
FROM items_meta
WHERE item_id = {{item_id}}
GROUP BY item_id, chain_id, item_name
ORDER BY count DESC
;

