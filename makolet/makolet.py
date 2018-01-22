"""Compares prices of items in the barcodes file."""
import argparse
import json
from typing import List, Dict

import prices
import stores

INPUT_FILE = "prices_dump.csv"


def pretty(val):
    """Returns a pretty representation of val, for human-readable printing."""
    return json.dumps(val, indent=2)


def prices_by_store(raw_data:
                    List[Dict[str, str]]) -> Dict[str, Dict[str, float]]:
    """Returns prices keyed by store, item name."""
    result = {}
    for d in raw_data:
        store = d["store_id"]
        item = d["item_name"]
        price = float(d["price"])
        if store not in result:
            result[store] = {}
        if item in result[store]:
            continue
        result[store][item] = price
    return result


parser = argparse.ArgumentParser()
parser.add_argument("db", help="SQLite database file to query.")
args = parser.parse_args()

barcodes = json.load(open("barcodes.json"))
data = [d for d in prices.get_prices_all(barcodes, args.db)]
print(len(data), "entries read.")
print("Distinct values:")
print(pretty({key: len({d[key] for d in data}) for key in data[0].keys()}))

print()
print("Number of stores having each item:")
items = {d["item_name"] for d in data}
print(pretty({p: len({d["store_id"] for d in data
                      if d["item_name"] == p}) for p in items}))

print()
print("Stores having all items:")
prices = prices_by_store(data)
good_stores = [s for s in prices if len(prices[s]) == len(items)]
print(good_stores)
print(len(good_stores), "stores.")

print()
print("Basket prices by store:")
store_names = {e[0]: e[1][0]["store_name"] for e in
               stores.get_stores(args.db).items()}
basket_prices = {store_names.get(s, s): sum(
    prices[s][i] for i in prices[s]) for s in good_stores}
print(pretty(basket_prices))

