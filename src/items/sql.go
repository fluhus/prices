package main

// Handles transforming data into SQL queries.

import (
	"fmt"
	"bytes"
	"strings"
)

// Insert commands should be performed in batches, since there is a limit
// on the maximal command length in SQLite.
const batchSize = 500

// Takes parsed entries from the XMLs and generates SQL queries for them.
// The time argument is used for creating timestamps. It should hold the time
// the data was published, in seconds since 1/1/1970 (Unix time).
type sqler func(data []map[string]string, time int64) []byte

// All available sqlers.
var sqlers = map[string]sqler {
	"stores": storeSqler,
	"prices": priceSqler,
}

// Creates SQL statements for stores.
func storeSqler(data []map[string]string, time int64) []byte {
	buf := bytes.NewBuffer(nil)
	data = escapeQuotes(data)
	
	// Insert into stores_id.
	fmt.Fprintf(buf, "INSERT OR IGNORE INTO stores_id VALUES\n")
	for i := range data {
		if i > 0 {
			fmt.Fprintf(buf, ",")
		}
		fmt.Fprintf(buf, "(NULL,'%s','%s','%s')\n", data[i]["chain_id"],
				data[i]["subchain_id"], data[i]["store_id"])
	}
	fmt.Fprintf(buf, ";\n")
	
	fmt.Fprintf(buf, "INSERT INTO stores_meta VALUES\n")
	for i, d := range data {
		if i > 0 {
			fmt.Fprintf(buf, ",")
		}
		fmt.Fprintf(buf, "(%d,(SELECT id FROM stores_id WHERE chain_id='%s'" +
				" AND subchain_id='%s' AND store_id='%s'),%s,%s,'%s','%s'," +
				"'%s','%s','%s','%s','%s','%s')\n",
				time,
				d["chain_id"], d["subchain_id"], d["store_id"], d["bikoret_no"],
				d["store_type"], d["chain_name"], d["subchain_name"],
				d["store_name"], d["address"], d["city"], d["zip_code"],
				d["last_update_date"], d["last_update_time"])
	}
	fmt.Fprintf(buf, ";\n")
	
	return buf.Bytes()
}

// Creates SQL statements for prices.
func priceSqler(data []map[string]string, time int64) []byte {
	buf := bytes.NewBuffer(nil)
	data = escapeQuotes(data)
	
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
				fmt.Fprintf(buf, "(NULL,'%s','%s','%s')\n",
						data[j]["item_type"], data[j]["item_code"],
						data[j]["chain_id"])
			} else {
				fmt.Fprintf(buf, "(NULL,'%s','%s','')\n", data[j]["item_type"],
						data[j]["item_code"])
			}
		}
		fmt.Fprintf(buf, ";\n")
	}
	
	return buf.Bytes()
}

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

