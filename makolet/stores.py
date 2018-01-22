"""An interface for fetching store information from the database."""
import sqlite

_QUERY = "SELECT * FROM stores_meta ORDER BY timestamp DESC;"


def get_stores(dbfile):
    """Returns a dict from store ID to its metadata."""
    result = {}
    for d in sqlite.query(dbfile, _QUERY):
        sid = d["store_id"]
        if sid not in result:
            result[sid] = []
        result[sid].append(d)
    return result

