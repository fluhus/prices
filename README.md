Israeli Supermarket Price Project
=================================

This is a collection of tools and scripts to download and manage price data
from Israeli vendors.

1. [Requirements](#requirements)
2. [How to Build](#how-to-build)
3. [How to Use](#how-to-use)
4. [Contributors](#contributors)

Requirements
------------

1. [Go](http://golang.org/) compiler added to your system path, to compile the
   go code.
2. Any databse program ([SQLite](http://sqlite.org/) recommended).

How to Build
------------

**No need to clone this repository yourself.**

1. Update GOPATH:
   * **Windows (cmd)** - `set GOPATH=\path\to\your\project`
   * **Linux (bash)** - `export GOPATH=/path/to/your/project`
2. Download the code: `go get github.com/fluhus/prices/...`
3. A `bin` folder will be created in the project's folder, containing all
   generated binaries.

How to Use
----------

### Downloading Price Data

1. Run `bin/scrape`. The program downloads all files from all known vendors.

### Creating the Database

1. Run `bin/parse`. The program parses XML files and outputs tab-separated
   text files.
2. Import the generated files to your database.

Contributors
------------

This project started at the Hebrew University of Jerusalem, carried out by
Amit Lavon, under the supervision of Dr. Aviv Zohar and Prof. Noam Nissan.
Now it is maintained independently by Amit Lavon.

Lists are ordered alphabetically.

### Coding

* Amit Lavon

### Schema Design

* Amit Lavon
* Dr. Aviv Zohar
* Ayelet Sapirstein
* Dr. Dafna Shahaf
* Gali Noti
* Prof. Noam Nissan
* Prof. Sara Cohen
* Yoni Sidi

### Testing & Data Curation

* Amit Lavon
* Asaf Kott
* Ayelet Sapirstein
* Dr. Dafna Shahaf
* Emanuel Marcu
* Gali Noti
* Yoni Sidi
