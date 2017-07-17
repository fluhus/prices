Israeli Supermarket Price Project
=================================

This is a collection of tools and scripts to download and manage price data
from Israeli vendors.

1. [License](#license)
2. [Requirements](#requirements)
3. [How to Build](#how-to-build)
4. [How to Use](#how-to-use)
5. [Contributors](#contributors)

License
-------

Copyright (c) 2015-2017 Amit Lavon

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

A link to the Github project (https://github.com/fluhus/prices) shall be
included in all copies or substantial portions of the Software. Any product
based on the Software must present the link either in its welcome, about or
help section, explicitly declaring that it is based on the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

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
