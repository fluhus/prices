Contribution
============

Thank you for your interest in contributing to the Prices project.
We welcome developers, analysts, designers, and anyone who wishes to join.

1. [What we Need Help With](#what-we-need-help-with)
1. [Recommended Background](#recommended-background)
1. [Technical How To's](#technical-how-tos)
1. [Contribution Guidelines](#contribution-guidelines)
1. [Contributors](#contributors)

What we Need Help With
----------------------

### Anything you think might be interesting. Let us know!

### Data

1. Help clean up the data and make it easier to process, for apps and further analysis.
2. Analyse the data to figure out trends and interesting things that we can publish.
3. Create cool visualizations that we can publish on the media.
4. Help us fetch old data from institutes who have it (HUJI, BOI..).

### Front End

1. Specifying the product - a site / app that would enable price comparison for consumers.
2. Assembling and leading the front-end team.

### Back End (Go Language)

1. Adding more scrapers as new chains join.
2. Help with maintaining the database, scraping and parsing routines.

Recommended Background
----------------------

* For data analysis and cleanup, knowledge in SQL, basic statistics.
* For frontend, knowledge of client and server languages - javascript, python/java/go.
* For backend coding, knowledge in the Go language, understanding HTTP.
* For backend maintenance, proficiency Linux and Bash, familiarity with Postgres.

Technical How To's
------------------

### Requirements

1. [Go](http://golang.org/) compiler added to your system path, to compile the go code.
2. Any databse program ([SQLite](http://sqlite.org/) recommended).

### How to Build

**No need to clone this repository yourself.**

1. Update GOPATH:
   * **Windows (cmd)** - `set GOPATH=\path\to\your\project`
   * **Linux (bash)** - `export GOPATH=/path/to/your/project`
2. Download the code: `go get github.com/fluhus/prices/...`
3. A `bin` folder will be created in the project's folder, containing all generated binaries.

### How to Use

#### Downloading Price Data

1. Run `bin/scrape`. The program downloads all files from all known vendors.

#### Creating the Database

1. Run `bin/parse`. The program parses XML files and outputs tab-separated text files.
2. Import the generated files to your database.

Contribution Guidelines
-----------------------

> TODO

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
