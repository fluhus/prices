package main

// Concrete parsers for parsing XML files.

// Maps a textual parser type to a parser.
var parsers = map[string]*parser{
	"prices": pricesParser,
	"stores": storesParser,
	"promos": promosParser,
}

// Parses price files.
var pricesParser = &parser{
	newCapturer("", "Item", "Product"),
	newCapturers(
		":chain_id", "ChainId",
		":subchain_id", "SubchainId",
		":store_id", "StoreId",
	),
	newCapturers(
		":update_time", "PriceUpdateDate",
		":item_code", "ItemCode",
		":item_name", "ItemName",
		":price", "ItemPrice",
		":item_type", "ItemType",
	),
	newCapturers(
		":manufacturer_name", "ManufacturerName",
		":manufacturer_country", "ManufacturerCountry",
		":manufacturer_item_description", "ManufacturerItemDescription",
		":unit_quantity", "UnitQty",
		":quantity", "Quantity",
		":unit_of_measure", "UnitOfMeasure",
		":is_weighted", "bIsWeighted", "blsWeighted",
		":quantity_in_package", "QtyInPackage",
		":unit_of_measure_price", "UnitOfMeasurePrice",
		":allow_discount", "AllowDiscount",
		":item_status", "ItemStatus",
	),
	nil,
}

// Parses store files.
var storesParser = &parser{
	newCapturer("", "Store"),
	newCapturers(
		":chain_id", "ChainId",
	),
	newCapturers(
		":subchain_id", "SubchainId",
		":store_id", "StoreId",
		":bikoret_no", "BikoretNo",
		":store_type", "StoreType",
		":chain_name", "ChainName",
		":subchain_name", "SubchainName",
		":store_name", "StoreName",
	),
	newCapturers(
		":address", "Address",
		":city", "City",
		":zip_code", "ZipCode",
		":last_update_time", "LastUpdateTime",
		":last_update_date", "LastUpdateDate",
	),
	nil,
}

// Parses promo files.
var promosParser = &parser{
	newCapturer("", "Promotion", "Sale"),
	newCapturers(
		":chain_id", "ChainId",
		":subchain_id", "SubchainId",
		":store_id", "StoreId",
	),
	newCapturers(
		":promotion_id", "PromotionId",
		":promotion_description", "PromotionDescription",
	),
	newCapturers(
		":promotion_start_date", "PromotionStartDate",
		":promotion_start_hour", "PromotionStartHour",
		":promotion_end_date", "PromotionEndDate",
		":promotion_end_hour", "PromotionEndHour",
		":reward_type", "RewardType",
		":allow_multiple_discounts", "AllowMultipleDiscounts",
		":min_qty", "MinQty",
		":max_qty", "MaxQty",
		":discount_rate", "DiscountRate",
		":discount_type", "DiscountType",
		":min_purchase_amnt", "MinPurchaseAmnt",
		":min_no_of_item_offered", "MinNoOfItemOfered",
		":price_update_date", "PriceUpdateDate",
		":discounted_price", "DiscountedPrice",
		":discounted_price_per_mida", "DiscountedPricePerMida",
		":additional_is_coupn", "AdditionalIsCoupon", "AdditionalsCoupon",
		":additional_gift_count", "AdditionalGiftCount",
		":additional_is_total", "AdditionalIsTotal",
		":additional_min_basket_amount", "AdditionalMinBasketAmount",
		":remarks", "Remarks",
	),
	newCapturers(
		":item_code", "ItemCode", "ItemId",
		":item_type", "ItemType",
		":is_gift_item", "IsGiftItem",
	),
}
