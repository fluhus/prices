package main

// Concrete parsers for parsing XML files.

// Maps a textual parser type to a parser.
var parsers = map[string]*parser {
	"prices": pricesParser,
	"stores": storesParser,
}

// Parsec price files.
var pricesParser = &parser {
	newXmlCapturer("(?:Item|Product)", ""),
	newXmlCapturers(
		"ChainId", "chain_id",
		"SubchainId", "subchain_id",
		"StoreId", "store_id",
	),
	newXmlCapturers(
		"PriceUpdateDate", "update_time",
		"ItemCode", "item_code",
		"ItemName", "item_name", 
		"ItemPrice", "price",
		"ItemType","item_type",
	),
	newXmlCapturers(
		"ManufacturerName","manufacturer_name",
		"ManufacturerCountry","manufacturer_country",
		"ManufacturerItemDescription","manufacturer_item_description",
		"UnitQty","unit_quantity",
		"Quantity","quantity",
		"UnitOfMeasure","unit_of_measure",
		"b(?:I|l)sWeighted","is_weighted",
		"QtyInPackage","quantity_in_package",
		"UnitOfMeasurePrice","unit_of_measure_price",
		"AllowDiscount","allow_discount",
		"ItemStatus","item_status",
	),
}

var storesParser = &parser {
	newXmlCapturer("Store", ""),
	newXmlCapturers(
		"ChainId", "chain_id",
	),
	newXmlCapturers(
		"SubchainId", "subchain_id",
		"StoreId", "store_id",
		"BikoretNo", "bikoret_no",
		"StoreType", "store_type",
		"ChainName", "chain_name",
		"SubchainName", "subchain_name",
		"StoreName", "store_name",
	),
	newXmlCapturers(
		"Address", "address",
		"City", "city",
		"ZipCode", "zip_code",
		"LastUpdateTime", "last_update_time",
		"LastUpdateDate", "last_update_date",
	),
}

