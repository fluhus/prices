SELECT item_id, count(*) count
FROM
(
    SELECT DISTINCT item_id, chain_id
    FROM items_meta
    WHERE
    manufacturer_item_description LIKE '%{{keyword}}%'
)
GROUP BY item_id
ORDER BY count DESC
LIMIT {{limit}}
;

