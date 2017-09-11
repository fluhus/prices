package main

// Reporting layer; converts field-maps to table entries.

import (
	"sort"
	"strings"

	"github.com/fluhus/prices/parse/bouncer"
)

// ----- REPORTER TYPE ---------------------------------------------------------

// Takes parsed entries from the XMLs and reports them to the bouncer.
// The time argument is used for creating timestamps. It should hold the time
// the data was published, in seconds since 1/1/1970 (Unix time).
type reporter func(data []map[string]string, time int64)

// TODO(amit): Make all reporters receive a single map.

// All available reporters.
var reporters = map[string]reporter{
	"stores": storesReporter,
	"prices": pricesReporter,
	"promos": promosReporter,
}

// ----- CONCRETE REPORTERS ----------------------------------------------------

// Reporter for stores.
func storesReporter(data []map[string]string, time int64) {
	// Get store-ids.
	ss := make([]*bouncer.Store, len(data))
	for i, d := range data {
		ss[i] = &bouncer.Store{
			d["chain_id"],
			d["subchain_id"],
			d["store_id"],
		}
	}
	sids := bouncer.MakeStoreIds(ss)

	// Report store-metas.
	metas := make([]*bouncer.StoreMeta, len(data))
	for i, d := range data {
		metas[i] = &bouncer.StoreMeta{
			time,
			sids[i],
			d["bikoret_no"],
			d["store_type"],
			d["chain_name"],
			d["subchain_name"],
			d["store_name"],
			d["address"],
			d["city"],
			d["zip_code"],
			d["last_update_date"],
			d["last_update_time"],
		}
	}

	bouncer.ReportStoreMetas(metas)
}

// Reporter for prices.
func pricesReporter(data []map[string]string, time int64) {
	// Report stores (just to get ids).
	ss := make([]*bouncer.Store, len(data))
	for i, d := range data {
		ss[i] = &bouncer.Store{
			d["chain_id"],
			d["subchain_id"],
			d["store_id"],
		}
	}
	sids := bouncer.MakeStoreIds(ss)

	// Report items.
	is := make([]*bouncer.Item, len(data))
	for i, d := range data {
		is[i] = &bouncer.Item{d["item_type"], d["item_code"], d["chain_id"]}
		if is[i].ItemType != "0" {
			is[i].ItemType = "1"
			is[i].ChainId = ""
		}
	}
	ids := bouncer.MakeItemIds(is)

	// Report item-metas.
	metas := make([]*bouncer.ItemMeta, len(data))
	for i, d := range data {
		metas[i] = &bouncer.ItemMeta{
			time,
			ids[i],
			d["chain_id"],
			d["update_time"],
			d["item_name"],
			d["manufacturer_item_description"],
			d["unit_quantity"],
			d["is_weighted"],
			d["quantity_in_package"],
			d["allow_discount"],
			d["item_status"],
		}
	}

	bouncer.ReportItemMetas(metas)

	// Report prices.
	prices := make([]*bouncer.Price, len(data))
	for i, d := range data {
		prices[i] = &bouncer.Price{
			time,
			ids[i],
			sids[i],
			d["price"],
			d["unit_of_measure_price"],
			d["unit_of_measure"],
			d["quantity"],
		}
	}

	bouncer.ReportPrices(prices)
}

// Reporter for promos.
func promosReporter(data []map[string]string, time int64) {
	// Get store id.
	sid := bouncer.MakeStoreIds([]*bouncer.Store{&bouncer.Store{
		data[0]["chain_id"],
		data[0]["subchain_id"],
		data[0]["store_id"],
	}})[0]

	// Create promos.
	promos := make([]*bouncer.Promo, len(data))
	for i, d := range data {
		promos[i] = &bouncer.Promo{
			time,
			d["chain_id"],
			d["promotion_id"],
			d["promotion_description"],
			d["promotion_start_date"],
			d["promotion_start_hour"],
			d["promotion_end_date"],
			d["promotion_end_hour"],
			d["reward_type"],
			d["allow_multiple_discounts"],
			d["min_qty"],
			d["max_qty"],
			d["discount_rate"],
			d["discount_type"],
			d["min_purchase_amnt"],
			d["min_no_of_item_offered"],
			d["price_update_date"],
			d["discounted_price"],
			d["discounted_price_per_mida"],
			d["additional_is_coupn"],
			d["additional_gift_count"],
			d["additional_is_total"],
			d["additional_min_basket_amount"],
			d["remarks"],
			sid,
			nil,
			nil,
		}
	}

	// Create item ids.
	for i, d := range data {
		// Get repeated fields.
		codes := strings.Split(d["item_code"], ";")
		types := strings.Split(d["item_type"], ";")
		gifts := strings.Split(d["is_gift_item"], ";")

		// Check lengths are all equal.
		if len(codes) != len(types) {
			// TODO(amit): Return an error.
			pe("Promo ignored promo due to mismatching lengths:", len(codes),
				len(types))
			continue
		}

		// Generate items.
		items := make([]*bouncer.Item, len(codes))
		for j := range codes {
			items[j] = &bouncer.Item{types[j], codes[j], promos[i].ChainId}
			if items[j].ItemType != "0" {
				items[j].ItemType = "1"
				items[j].ChainId = ""
			}
		}
		promos[i].ItemIds = bouncer.MakeItemIds(items)

		// Generate gift items.
		promos[i].GiftItems = make([]string, len(codes))
		if len(gifts) == len(codes) {
			for j := range gifts {
				promos[i].GiftItems[j] = gifts[j]
			}
		}

		sort.Sort(&itemsAndGifts{promos[i].ItemIds, promos[i].GiftItems})
	}

	bouncer.ReportPromos(promos)
}

// ----- OTHER HELPERS ---------------------------------------------------------

// Sorting interface for item-IDs and corresponding gift field.
type itemsAndGifts struct {
	items []int
	gifts []string
}

func (iag *itemsAndGifts) Len() int {
	return len(iag.items)
}

func (iag *itemsAndGifts) Less(i, j int) bool {
	return iag.items[i] < iag.items[j]
}

func (iag *itemsAndGifts) Swap(i, j int) {
	iag.items[i], iag.items[j] = iag.items[j], iag.items[i]
	iag.gifts[i], iag.gifts[j] = iag.gifts[j], iag.gifts[i]
}
