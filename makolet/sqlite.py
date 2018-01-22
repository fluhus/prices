"""A simple interface for querying an SQLite database."""
import csv
import subprocess as sp
from os.path import isfile
from typing import Dict, Iterable

_SQL_PRE_QUERY = ".headers on\n.mode csv\n"


def query(dbfile: str, sqlquery: str) -> Iterable[Dict[str, str]]:
    """Executes a query in the given SQLite file.

    Returns:
        A dict for each line, from column name to value.
    """
    if not isfile(dbfile):
        raise Exception("file not found: " + dbfile)

    # FIXME: This is an unsafe way to execute SQL queries, which is vulnerable
    # to SQL injection. We should work with a dedicated sql package instead.
    process = sp.Popen(["sqlite3", dbfile], stdout=sp.PIPE, stdin=sp.PIPE)
    process.stdin.write(_SQL_PRE_QUERY.encode())
    process.stdin.write(sqlquery.encode())
    process.stdin.close()
    return csv.DictReader(line.decode().strip() for line in process.stdout)

