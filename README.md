Israeli Supermarket Price Project
=================================

This is a collection of tools and scripts to download and manage price data
from Israeli vendors.

1. [License](#license)
2. [Requirements](#requirements)
3. [How to Build](#how-to-build)
4. [How to Use](#how-to-use)

License
-------

Copyright (c) 2015-2016 Amit Lavon

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

1. Update GOPATH:
   * **Linux** - `export GOPATH=/path/to/project`
   * **Windows** - `set GOPATH=\path\to\project`
2. Compile everything. `cd` into project/src folder and execute:
   `go install ./...`
3. A `bin` folder will be created in the project's folder, containing all
   generated binaries.

How to Use
----------

### Downloading Price Data

1. Build the executables.
2. Run `bin/prices`. The program downloads all files from all known vendors.

### Creating the Database

1. Build the executables.
2. Run `bin/items`. The program parses XML filess and outputs tab-separated
   text files.
3. Import the generated files to your database.


