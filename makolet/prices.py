"""An interface for fetching prices from the database."""
from typing import Iterable, Dict, List

import sqlite

_QUERY_FILE = "prices_lookup.sql"
_SQL_PRICE_QUERY = open(_QUERY_FILE).read()


def get_prices(barcode: str, dbfile: str) -> Iterable[Dict]:
    """Returns prices for the given barcode as dicts.

    Each element is a dictionary that represents a price in a certain store at
    a certain time.
    """
    query = _SQL_PRICE_QUERY.replace("{{BARCODE}}", barcode)
    return sqlite.query(dbfile, query)


def get_prices_all(barcodes: Dict[str, List[str]],
                   dbfile: str) -> Iterable[Dict]:
    """Returns prices for the given barcodes as dicts.

    Each element is a dictionary that represents a price in a certain store at
    a certain time.
    """
    for item in barcodes:
        for barcode in barcodes[item]:
            for obj in get_prices(barcode, dbfile):
                obj["item_name"] = item
                yield obj

