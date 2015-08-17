package main

// Handles transforming parsed data into SQL queries.

import (
	"fmt"
	"bytes"
	"strings"
	"hash/crc64"
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
	buf := bytes.NewBuffer(nil)
	data = escapeQuotes(data)
	
	// Insert into stores_id.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO stores_id VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			fmt.Fprintf(buf, "(NULL,'%s','%s','%s')\n", data[j]["chain_id"],
					data[j]["subchain_id"], data[j]["store_id"])
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	// Insert into stores_meta.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT INTO stores_meta VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			fmt.Fprintf(buf, "(%d,(SELECT id FROM stores_id WHERE chain_id=" +
					"'%s' AND subchain_id='%s' AND store_id='%s'),%s,%s,'%s'," +
					"'%s','%s','%s','%s','%s','%s','%s')\n",
					time, data[j]["chain_id"], data[j]["subchain_id"],
					data[j]["store_id"], data[j]["bikoret_no"],
					data[j]["store_type"], data[j]["chain_name"],
					data[j]["subchain_name"], data[j]["store_name"],
					data[j]["address"], data[j]["city"], data[j]["zip_code"],
					data[j]["last_update_date"], data[j]["last_update_time"])
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	return buf.Bytes()
}

// Creates SQL statements for prices.
func pricesSqler(data []map[string]string, time int64) []byte {
	buf := bytes.NewBuffer(nil)
	data = escapeQuotes(data)
	
	// Make sure store is in stores_id (sometimes it isn't on the stores file).
	if len(data) > 0 {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO stores_id VALUES (NULL,'%s'" +
				",'%s','%s');\n", data[0]["chain_id"], data[0]["subchain_id"],
				data[0]["store_id"])
	}
	
	// Insert into items_id.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO items_id VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			
			// Item-type=0 means an internal code, hence we identify by chain
			// ID.
			if data[j]["item_type"] == "0" {
				fmt.Fprintf(buf, "(NULL,0,'%s','%s')\n",
						data[j]["item_code"], data[j]["chain_id"])
			} else {
				fmt.Fprintf(buf, "(NULL,1,'%s','')\n", data[j]["item_code"])
			}
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	// Insert into items_meta.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT INTO items_meta VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			
			// Items of type 0 (internal barcode) are identified with chain_id.
			selectItem := ""
			if data[j]["item_type"] == "0" {
				selectItem = "SELECT id FROM items_id WHERE item_type=0 AND " +
						"item_code='" + data[j]["item_code"] + "' AND " +
						"chain_id='" + data[j]["chain_id"] + "'"
			} else {
				selectItem = "SELECT id FROM items_id WHERE item_type=1 AND " +
						"item_code='" + data[j]["item_code"] + "' AND " +
						"chain_id=''"
			}
			
			fmt.Fprintf(buf, "(%d,(%s),'%s','%s','%s','%s','%s','%s','%s'," +
					"'%s','%s','%s','%s',%d)\n", time, selectItem,
					data[j]["chain_id"],
					data[j]["update_time"],
					data[j]["item_name"],
					data[j]["manufacturer_item_description"],
					data[j]["unit_quantity"], data[j]["quantity"],
					data[j]["unit_of_measure"], data[j]["is_weighted"],
					data[j]["quantity_in_package"],
					data[j]["allow_discount"], data[j]["item_status"],
					rowCrc(data[j], itemsMetaCrc))
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	// Insert into prices.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT INTO prices VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			
			// Items of type 0 (internal barcode) are identified with chain_id.
			selectItem := ""
			if data[j]["item_type"] == "0" {
				selectItem = "SELECT id FROM items_id WHERE item_type=0 AND " +
						"item_code='" + data[j]["item_code"] + "' AND " +
						"chain_id='" + data[j]["chain_id"] + "'"
			} else {
				selectItem = "SELECT id FROM items_id WHERE item_type=1 AND " +
						"item_code='" + data[j]["item_code"] + "' AND " +
						"chain_id=''"
			}
			
			selectStore := "SELECT id FROM stores_id WHERE chain_id='" +
					data[j]["chain_id"] + "' AND subchain_id='" +
					data[j]["subchain_id"] + "' AND store_id='" +
					data[j]["store_id"] + "'"
			
			fmt.Fprintf(buf, "(%d,(%s),(%s),%s,%s)\n", time, selectItem,
					selectStore, data[j]["price"],
					data[j]["unit_of_measure_price"])
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	return buf.Bytes()
}

// Creates SQL statements for promos.
func promosSqler(data []map[string]string, time int64) []byte {
	// TODO(amit): This sqler is not ready!
	buf := bytes.NewBuffer(nil)
	data = escapeQuotes(data)
	
	// Make sure store is in stores_id (sometimes it isn't on the stores file).
	if len(data) > 0 {
		fmt.Fprintf(buf, "INSERT OR IGNORE INTO stores_id VALUES (NULL,'%s'" +
				",'%s','%s');\n", data[0]["chain_id"], data[0]["subchain_id"],
				data[0]["store_id"])
	}
	
	// Insert into promos.
	for i := 0; i < len(data); i += batchSize {
		fmt.Fprintf(buf, "INSERT INTO promos VALUES\n")
		for j := i; j < len(data) && j < i+batchSize; j++ {
			if j > i {
				fmt.Fprintf(buf, ",")
			}
			
			selectStore := "SELECT id FROM stores_id WHERE chain_id='" +
					data[j]["chain_id"] + "' AND subchain_id='" +
					data[j]["subchain_id"] + "' AND store_id='" +
					data[j]["store_id"] + "'"
			
			fmt.Fprintf(buf, "(NULL,%d,%d,(%s),'%s','%s','%s','%s','%s','%s'," +
					"'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s'," +
					"'%s','%s','%s','%s','%s',%d)",
					time,
					time + (60*60*24), // Add one day, so that the promo
					                   // holds until tomorrow.
					selectStore,
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
					rowCrc(data[j], promosCrc))
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	return buf.Bytes()
}


// ----- HELPERS ---------------------------------------------------------------

// Escapes quotation characters in parsed data. Input data is unchanged.
func escapeQuotes(maps []map[string]string) []map[string]string {
	result := make([]map[string]string, len(maps))
	for i := range maps {
		result[i] = map[string]string{}
		for k, v := range maps[i] {
			result[i][k] = strings.Replace(v, "'", "''", -1)
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
	"quantity",
	"unit_of_measure",
	"is_weighted",
	"quantity_in_package",
	"allow_discount",
	"item_status",
}

// Fields to include in CRC for promos table.
var promosCrc = []string {
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
}

