"""An interface for fetching item metadata from the database."""
from typing import Iterable, Dict, List

import sqlite

_FETCH_ID_FILE = "items_fetch_id.sql"
_FETCH_ID_QUERY = open(_FETCH_ID_FILE).read()
_FETCH_NAMES_FILE = "items_fetch_names.sql"
_FETCH_NAMES_QUERY = open(_FETCH_NAMES_FILE).read()


def get_ids(keyword: str, limit: int,
            dbfile: str) -> Iterable[Dict[str, str]]:
    """Returns entries with IDs of items containing the given keyword."""
    query = _FETCH_ID_QUERY.replace("{{keyword}}", keyword).replace(
        "{{limit}}", str(limit))
    return sqlite.query(dbfile, query)


def get_names(item_id: str, dbfile: str) -> List[str]:
    """Returns a list of names for the given item ID."""
    query = _FETCH_NAMES_QUERY.replace("{{item_id}}", str(item_id))
    chain_ids = set()
    names = []
    for e in sqlite.query(dbfile, query):
        if e["chain_id"] in chain_ids:
            continue
        chain_ids |= {e["chain_id"]}
        names.append(e["item_name"])
    return names

