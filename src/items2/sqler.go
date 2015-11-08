package main

// Handles transforming of field maps into SQL queries.

import (
	"fmt"
	"bytes"
	"strings"
	"hash/crc64"
	"bouncer"
)


// ----- SQLER TYPE ------------------------------------------------------------

// Takes parsed entries from the XMLs and generates SQL queries for them.
// The time argument is used for creating timestamps. It should hold the time
// the data was published, in seconds since 1/1/1970 (Unix time).
type sqler func(data []map[string]string, time int64) []byte

// All available sqlers.
var sqlers = map[string]sqler {
	"stores": storesSqler,
	"prices": pricesSqler,
	"promos": promosSqler,
}


// ----- CONCRETE SQLERS -------------------------------------------------------

// Insert commands should be performed in batches, since there is a limit
// on the maximal insert size in SQLite.
const batchSize = 500

// Creates SQL statements for stores.
func storesSqler(data []map[string]string, time int64) []byte {
	data = escapeQuotes(data)
	
	// Get store-ids.
	ss := make([]*bouncer.Store, len(data))
	for i, d := range data {
		ss[i] = &bouncer.Store {
			d["chain_id"],
			d["subchain_id"],
			d["store_id"],
		}
	}
	sids := bouncer.MakeStoreIds(ss)
	
	// Report store-metas.
	metas := make([]*bouncer.StoreMeta, len(data))
	for i, d := range data {
		metas[i] = &bouncer.StoreMeta {
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
	
	return nil
}

// Creates SQL statements for prices.
func pricesSqler(data []map[string]string, time int64) []byte {
	data = escapeQuotes(data)

	// Report stores (just to get ids).
	ss := make([]*bouncer.Store, len(data))
	for i, d := range data {
		ss[i] = &bouncer.Store {
			d["chain_id"],
			d["subchain_id"],
			d["store_id"],
		}
	}
	sids := bouncer.MakeStoreIds(ss)

	// Report items.
	is := make([]*bouncer.Item, len(data))
	for i, d := range data {
		is[i] = &bouncer.Item {d["item_type"], d["item_code"], d["chain_id"]}
		if is[i].ItemType != `"0"` {
			is[i].ItemType = `"1"`
			is[i].ChainId = ""
		}
	}
	ids := bouncer.MakeItemIds(is)
	
	// Report item-metas.
	metas := make([]*bouncer.ItemMeta, len(data))
	for i, d := range data {
		metas[i] = &bouncer.ItemMeta {
			time,
			ids[i],
			sids[i],
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
		prices[i] = &bouncer.Price {
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
	
	return nil
}

// Creates SQL statements for promos.
func promosSqler(data []map[string]string, time int64) []byte {
	buf := bytes.NewBuffer(nil)
	data = escapeQuotes(data)
	
	// Make sure store is in stores (sometimes it isn't on the stores file).
	if len(data) > 0 {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO stores VALUES (NULL,'%s'" +
				",'%s','%s');\n", data[0]["chain_id"], data[0]["subchain_id"],
				data[0]["store_id"])
	}
	
	// Compute CRC for each promo.
	for i := range data {
		data[i]["crc"] = fmt.Sprint(rowCrc(data[i], promosCrc))
	}
	
	// Promos with more items than that are not reported in promos_items.
	const maxItemsInPromo = 100
	
	// Items counts for each promo are stored here, to be reported in the promos
	// table.
	numberOfItems := make([]int, len(data))
	
	// Generate an array of items.
	type promoItem struct {
		code    string
		typ     string
		gift    string
		chain   string
		crc     string
		promoId string
	}
	
	items := []*promoItem{}
	for i := range data {
		// Get repeated fields.
		codes := strings.Split(data[i]["item_code"], ";")
		types := strings.Split(data[i]["item_type"], ";")
		gifts := strings.Split(data[i]["is_gift_item"], ";")
		
		// Check lengths are all equal.
		if len(codes) != len(types) {
			// TODO(amit): Return an error.
			pe("Promo ignored promo due to mismatching lengths:", len(codes),
					len(types))
			continue
		}
		
		numberOfItems[i] = len(codes)

		// Ignore promo if it has too many items.
		if len(codes) > maxItemsInPromo {
			continue
		}
		
		// Generate items.
		for j := range codes {
			if len(gifts) == len(codes) {
				items = append(items, &promoItem{codes[j], types[j], gifts[j],
						data[i]["chain_id"], data[i]["crc"],
						data[i]["promotion_id"]})
			} else {
				items = append(items, &promoItem{codes[j], types[j], "0",
						data[i]["chain_id"], data[i]["crc"],
						data[i]["promotion_id"]})
			}
		}
	}
	
	// Insert into promos.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT INTO promos VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			
			tooMany := 0
			if numberOfItems[j] > maxItemsInPromo {
				tooMany = 1;
			}
			
			fmt.Fprintf(buf, "(NULL,%d,%d,(%s),'%s','%s','%s','%s','%s','%s'," +
					"'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s'," +
					"'%s','%s','%s','%s','%s',%d,%d,%s)\n",
					time,
					time + (60*60*24), // Add one day, so that the promo
					                   // holds until tomorrow.
					data[j]["chain_id"],
					data[j]["reward_type"],
					data[j]["allow_multiple_discounts"],
					data[j]["promotion_id"],
					data[j]["promotion_description"],
					data[j]["promotion_start_date"],
					data[j]["promotion_start_hour"],
					data[j]["promotion_end_date"],
					data[j]["promotion_end_hour"],
					data[j]["min_qty"],
					data[j]["max_qty"],
					data[j]["discount_rate"], 
					data[j]["discount_type"],
					data[j]["min_purchase_amnt"],
					data[j]["min_no_of_item_offered"],
					data[j]["price_update_date"],
					data[j]["discounted_price"],
					data[j]["discounted_price_per_mida"],
					data[j]["additional_is_coupon"],
					data[j]["additional_gift_count"],
					data[j]["additional_is_total"],
					data[j]["additional_min_basket_amount"],
					data[j]["remarks"],
					numberOfItems[j],
					tooMany,
					data[j]["crc"])
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	// Insert into promos_stores.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO promos_stores VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}

			selectPromo := "SELECT promo_id FROM promos " +
					"WHERE crc=" + data[j]["crc"] +
					" AND chain_id='" + data[j]["chain_id"] + "'" +
					" AND promotion_id='" + data[j]["promotion_id"] + "'"

			selectStore := "SELECT store_id FROM stores WHERE chain_id='" +
					data[j]["chain_id"] + "' AND subchain_id='" +
					data[j]["subchain_id"] + "' AND reported_store_id='" +
					data[j]["store_id"] + "'"
			
			fmt.Fprintf(buf, "((%s),(%s))\n", selectPromo, selectStore)
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	// Insert into items, in case an item is not present.
	for i := 0; i < len(items); i += batchSize {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO items VALUES\n")
		for j := i; j < len(items) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			
			// TODO(amitl): Extract this insert into items to a function.

			// Item-type=0 means an internal code, hence we identify by chain
			// ID.
			if items[j].typ == "0" {
				fmt.Fprintf(buf, "(NULL,0,'%s','%s')\n",
						items[j].code, items[j].chain)
			} else {
				fmt.Fprintf(buf, "(NULL,1,'%s','')\n", items[j].code)
			}
		}
		fmt.Fprintf(buf, ";\n")
	}

	// Insert into promos_items.
	for i := 0; i < len(items); i += batchSize {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO promos_items VALUES\n")
		for j := i; j < len(items) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			
			selectPromo := "SELECT promo_id FROM promos " +
					"WHERE crc=" + items[j].crc +
					" AND chain_id='" + items[i].chain + "'" +
					" AND promotion_id='" + items[i].promoId + "'"
			
			fmt.Fprintf(buf, "((%s),(%s),'%s')\n",
					selectPromo,
					selectItem(items[j].typ, items[j].code, items[j].chain),
					items[j].gift)
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	return buf.Bytes()
}


// ----- SQL HELPERS -----------------------------------------------------------

// Returns a SELECT statement for item-id. The statement has no parenthesis and
// no semicolon at the end.
func selectItem(typ, code, chain string) string {
	// Items of type 0 (internal barcode) are identified with chain_id.
	if typ == "0" {
		return "SELECT item_id FROM items WHERE item_type=0 AND " +
				"item_code='" + code + "' AND " +
				"chain_id='" + chain + "'"
	} else {
		return "SELECT item_id FROM items WHERE item_type=1 AND " +
				"item_code='" + code + "' AND " + "chain_id=''"
	}
}


// ----- OTHER HELPERS ---------------------------------------------------------

// Escapes quotation characters in parsed data. Input data is unchanged.
func escapeQuotes(maps []map[string]string) []map[string]string {
	result := make([]map[string]string, len(maps))
	for i := range maps {
		result[i] = map[string]string{}
		for k, v := range maps[i] {
			result[i][k] = "\"" + strings.Replace(v, "\"", "\"\"", -1) + "\""
		}
	}
	return result
}

// Calculates CRC from a given row's fields.
func rowCrc(data map[string]string, fields []string) int64 {
	crc := crc64.New(crcTable)
	
	for _, field := range fields {
		fmt.Fprint(crc, data[field])
	}
	
	return int64(crc.Sum64())
}

// Default table for CRC.
var crcTable = crc64.MakeTable(crc64.ECMA)

// Fields to include in CRC for items_meta table.
var itemsMetaCrc = []string {
	"item_name",
	"manufacturer_item_description",
	"unit_quantity",
	"is_weighted",
	"quantity_in_package",
	"allow_discount",
	"item_status",
}

// Fields to include in CRC for prices table.
var pricesCrc = []string {
	"price",
	"unit_of_measure_price",
	"unit_of_measure",
	"quantity",
}

// Fields to include in CRC for promos table.
var promosCrc = []string {
	"chain_id",
	"reward_type",
	"allow_multiple_discounts",
	"promotion_id",
	"promotion_description",
	"promotion_start_date",
	"promotion_start_hour",
	"promotion_end_date",
	"promotion_end_hour",
	"min_qty",
	"max_qty",
	"discount_rate",
	"discount_type",
	"min_purchase_amnt",
	"min_no_of_item_offered",
	"price_update_date",
	"discounted_price",
	"discounted_price_per_mida",
	"additional_is_coupn",
	"additional_gift_count",
	"additional_is_total",
	"additional_min_basket_amount",
	"remarks",
	"item_code",
	"item_type",
	"is_gift_item",
}

