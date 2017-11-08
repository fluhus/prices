General Information
===================

See XML specifications here: https://drive.google.com/file/d/0Bw2XXw9aHzlCT0xRMV9WSnZIS0E/view

See unresolved data issues here: https://docs.google.com/document/d/168h95u3p-Xx7hSAHj_qKhM2aIo2G3vkmI3qDB1z-bNI/

PLEASE BE WARNED: Vendors report garbage. Unless stated otherwise, any piece of information presented in this database is brought 'as is' and should be treated as unreliable, possibly incorrect, badly formatted, unsafe for use, offensive, and inappropriate for children.

Data fields that were created or curated by us are marked explicitly as safe.

## chains

Maps chain code to chain name.

**Fields**

* **chain_id:** Chain code, as provided by GS1. (safe)
* **chain_name:** Chain name in English. (safe)
## stores

Identifies every store in the data. Each store may appear once.

**Fields**

* **store_id:**  (safe)
* **chain_id:** Chain code, as provided by GS1.
* **subchain_id:** Subchain number.
* **reported_store_id:** Store number issued by the chain.
## stores_meta

Metadata about stores. Each store may appear several times.

**Fields**

* **timestamp:** Unix time when this entry was encountered. (safe)
* **store_id:** References stores.store_id. (safe)
* **bikoret_no:** ???
* **store_type:** 1 for physical, 2 for online, 3 for both.
* **chain_name:** 
* **subchain_name:** 
* **store_name:** 
* **address:** 
* **city:** 
* **zip_code:** 
* **last_update_date:** 
* **last_update_time:** 
## items

Identifies every commodity item in the data. Each item may appear once.

**Fields**

* **item_id:**  (safe)
* **item_type:** 0 for internal codes, 1 for barcodes.
* **item_code:** Barcode number or internal code.
* **chain_id:** Empty string for universal.
## items_meta

Contains all metadata about each item. Each item may appear several times.

**Fields**

* **timestamp:** Unix time when this entry was encountered. (safe)
* **item_id:** References items(item_id). (safe)
* **chain_id:** Chain code, as provided by GS1.
* **update_time:** 
* **item_name:** 
* **manufacturer_item_description:** 
* **unit_quantity:** 
* **is_weighted:** 1 if sold in bulk, 0 if not.
* **quantity_in_package:** Quantity of units in a package.
* **allow_discount:** Is the item allowed in promotions.
* **item_status:** ???
## prices

Contains all reported prices for all items.

**Fields**

* **timestamp:** Unix time when this entry was encountered. (safe)
* **item_id:** References items.item_id. (safe)
* **store_id:** References stores.store_id. (safe)
* **price:** Price in shekels as reported in raw data.
* **unit_of_measure_price:** Price in shekels as reported in raw data.
* **unit_of_measure:** Gram, liter, etc.
* **quantity:** How many grams/liters etc.
## promos

Identifies every promotion in the data. Promo id and metadata are saved together since they are unique. A change in the metadata will be registered as a new promo.

**Fields**

* **promo_id:**  (safe)
* **timestamp_from:** Unix time when this entry was first encountered. (safe)
* **timestamp_to:** Unix time when this entry was last encountered + one day. (safe)
* **chain_id:** Chain code, as provided by GS1.
* **promotion_id:** Issued by the chain, not by us.
* **promotion_description:** 
* **promotion_start_date:** 
* **promotion_start_hour:** 
* **promotion_end_date:** 
* **promotion_end_hour:** 
* **reward_type:** ???
* **allow_multiple_discounts:** 'Kefel mivtzaim'.
* **min_qty:** Min quantity for triggering promo.
* **max_qty:** Max quantity for triggering promo.
* **discount_rate:** 
* **discount_type:** 0 for relative, 1 for absolute.
* **min_purchase_amnt:** 
* **min_no_of_item_offered:** Like min_qty, not sure what the difference is.
* **price_update_date:** 
* **discounted_price:** 
* **discounted_price_per_mida:** 
* **additional_is_coupn:** 1 if depends on coupon, 0 if not.
* **additional_gift_count:** Number of gift items.
* **additional_is_total:** Promo is on all items in the store.
* **additional_min_basket_amount:** ???
* **remarks:** 
* **number_of_items:** Number of items that take part in the promotion. Should be equivalent to count(*) on the promo_id in promos_items, but some of the promos are not reported there. (safe)
* **not_in_promos_items:** 0 if reported in promos_items, 1 if not. (safe)
## promos_stores

Reports what stores take part in every promo. A single promo may have several rows, one for each store.

**Fields**

* **promo_id:** References promos.promo_id. (safe)
* **store_id:** References stores.store_id. (safe)
## promos_items

Reports what items take part in every promo. A single promo may have several rows, one for each item.

CAVEAT: promos that include more than 100 items are not reported here, because those promos usually apply on an entire store ("everything for 10% discount") and that bloats the DB. They are reported on the other tables as usual.

**Fields**

* **promo_id:** References promos.promo_id. (safe)
* **item_id:** References items.item_id. (safe)
* **is_gift_item:** 
