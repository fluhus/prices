package main

import (
	"fmt"
	"bytes"
	"strings"
)

// Takes parsed entries from the XMLs and generates SQL queries for them.
type sqler interface {
	toSql(data []map[string]string) []byte
}

// All available sqlers.
var sqlers = map[string]sqler {
	"stores": &storesSqler{},
}

// Creates SQL statements for stores.
type storesSqler struct{}

func (s *storesSqler) toSql(data []map[string]string) []byte {
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
		fmt.Fprintf(buf, "(0,(SELECT id FROM stores_id WHERE chain_id='%s'" +
				" AND subchain_id='%s' AND store_id='%s'),%s,%s,'%s','%s'," +
				"'%s','%s','%s','%s','%s','%s')\n",
				d["chain_id"], d["subchain_id"], d["store_id"], d["bikoret_no"],
				d["store_type"], d["chain_name"], d["subchain_name"],
				d["store_name"], d["address"], d["city"], d["zip_code"],
				d["last_update_date"], d["last_update_time"])
	}
	fmt.Fprintf(buf, ";\n")
	
	return buf.Bytes()
}

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

