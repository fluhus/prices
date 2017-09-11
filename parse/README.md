Parser - Design
===============

Objective
---------

This readme describes the general design of the parser.

It has 2 primary goals:

1. To keep a record of the decisions made in the development process, why we made them and what we learned.
2. To give contributors a better understanding of the program, so they can contribute more efficiently.

Components
----------

### Overview

```
(hard drive)
 |
 | Raw data files
 |
 V
Loader
 |
 | Decompressed XML text
 |
 V
Parser
 |
 | String-string maps, according to regulator's schema
 |
 V
Reporter
 |
 | Table entry objects, according to our schema
 |
 V
Bouncer
 |
 | Text tables
 |
 V
(hard drive)
```

### Loader

The loader provides an abstraction over the file system. Data files come in different formats - raw text, gzip or zip. The loader infers the file type and applies the appropriate decompression method.

Outputs the raw textual data of the files.

### Parser

The parser parses the xml into meaningful data structures. It has 2 steps: first it parses the raw text into a hierarchical node representation; then it traverses the node tree, extracts fields and values and puts them in maps.

The keys of the maps are the fields required by the Price Transparency Regulations.

For each file, outputs a list of string-string maps. Each item (product, promo, etc.) in the file is represented in a single map, from field name to its value. For example:

```json
{
  "item_name": "Coca Cola 330 ml",
  "price": "5.00",
  "...": "..."
}
```

#### Tolerating Variance

The parser is built in a way that allows it to tolerate typos and variance in field names. Due to ambiguity of the regulation and lack of enforcement, different chains name their fields in different ways. For example, some chains wrap each product with a `<product>` tag, and some with an `<item>` tag.

### Reporter

The reporter takes the string-string maps and generates data objects that correspond to rows in our database. The data rows generated here are ready to be written as-is to the database.

### Bouncer

The bouncer is the final layer between the data rows and the hard drive. The bouncer's goal is to filter out repeating pieces of data.

Price data is published every day but changes rarely, so there are many repetitions of the same data rows. The bouncer uses hash to keep track of what data items have already appeared.

#### Hashing

Every reported data row is hashed, and its hash is kept in a set. Before reporting a row, the bouncer first checks the hash-set and if the hash code exists, it skips that row.

This approach has a small downside, which is sensitivity to hash collisions. However, using a 64-bit hash size give a satisfyingly low collision probability (less than 3e-6 for 10M items).

#### Rejected approach: keeping the latest piece of data about each item in memory

Previously, the parser kept the latest piece of information about each item in memory. This way it could compare a new report about the item to the latest, and report only if there was a change.

For example, it would remember that the latest price for item 123 was 10.00. Upon a new report about item 123, if the price was still 10.00, it would ignore that report; if the price was 9.00, it would print the new price and update its in-memory data.

That approach was rejected because chains report different items under the same codes. So what would happen was that there were 2 real items that have code 123, and in the reports it would look like the price of item 123 alternates between 2 different prices daily. That would render the bouncer ineffective in bouncing off these repetitions.

Procedures
----------

> TODO
