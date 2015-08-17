package main

// Concrete parsers for parsing XML files.

// Maps a textual parser type to a parser.
var parsers = map[string]*parser {
	"prices": pricesParser,
	"stores": storesParser,
	"promos": promosParser,
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

var promosParser = &parser {
	newXmlCapturer("Promotion", ""),
	newXmlCapturers(
		"ChainId", "chain_id",
		"SubchainId", "subchain_id",
		"StoreId", "store_id",
	),
	newXmlCapturers(
		"RewardType", "reward_type",
		"AllowMultipleDiscounts", "allow_multiple_discounts",
		"PromotionId", "promotion_id",
		"PromotionDescription", "promotion_description",
		"PromotionStartDate", "promotion_start_date",
		"PromotionStartHour", "promotion_start_hour",
		"PromotionEndDate", "promotion_end_date",
		"PromotionEndHour", "promotion_end_hour",
		"MinQty", "min_qty",
		"MaxQty", "max_qty",
		"DiscountRate", "discount_rate",
		"DiscountType", "discount_type",
		"MinPurchaseAmnt", "min_purchase_amnt",
		"MinNoOfItemOfered", "min_no_of_item_offered",
	),
	newXmlCapturers(
		"PriceUpdateDate", "price_update_date",
		"DiscountedPrice", "discounted_price",
		"DiscountedPricePerMida", "discounted_price_per_mida",
		"AdditionalI?sCoupon", "additional_is_coupn",
		"AdditionalGiftCount", "additional_gift_count",
		"AdditionalIsTotal", "additional_is_total",
		"AdditionalMinBasketAmount", "additional_min_basket_amount",
		"Remarks", "remarks",
	),
}

