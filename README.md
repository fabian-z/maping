#maping - [![Build Status](https://travis-ci.org/nfdesign/maping.svg?branch=master)](https://travis-ci.org/nfdesign/maping)
mail account ping - utility for checking sets of mail servers (SMTP/IMAPv4).
Saves results to database and is able to generate an SVG data visualization matrix from the results.

#Features

* Cross-platform - targeting all platforms go targets
* No hard CGO dependency, although it needs the system x509 certificate store

---

* Implementing SMTP 8BITMIME, AUTH, STARTTLS (using net/smtp)
* Implementing IMAP4rev1 using [go-imap](https://github.com/mxk/go-imap) - including support for AUTHENTICATE with CRAM-MD5
* SMTPS and IMAPS support (Explicit SSL) - supports AUTH PLAIN/LOGIN over secure connection
* Supporting virtually any mail hoster offering IMAP access

---

* Fully configurable using JSON-like syntax (extended with comments)
* Automatic test pair generation using post-processed cartesian products
* Concurrent execution of tests using worker pool
* Integrated, pure go database using [ql](https://github.com/cznic/ql)
* Stores gzip compressed protocol logs alongside with test data in a separate database table

----

* Test data visualization in an SVG matrix (separate tool)
* Ability to visualize latest or specific dataset from given database

###To be done

* Unit tests
* godoc

#Getting started

You will need to have a working go environment installed. Refer to your distributions manual, the [go manual](https://golang.org/doc/install) or support forums on how to accomplish this.

```
go get github.com/nfdesign/maping
go get github.com/nfdesign/maping/gensvg
```

*Since we are using ql as database, you may need to build with tag "purego", if you want to avoid any CGO dependency while 
building. This disables some speed optimizations for the built-in database ql.
Please see [this issue](https://github.com/cznic/ql/issues/86).*

After this command, the executables maping and gensvg will be in $GOPATH/bin. Use the example,json configuration from this repository or from $GOPATH/src/github.com/nfdesign/maping/example.json to create a new configuration. By default, maping will look for config.json in its own directory and for /etc/maping.json. You may also specifiy a configuration file on the command line (see below).
After the first run, you may use the gensvg command to create a SVG visualization of the generated test data from the database like this:

```
$GOPATH/bin/gensvg -db="maping.db"
```

#Command usage reference

```
$ gensvg --help
Usage of gensvg:
  -db="maping.db": Path to database
  -output="output.svg": Output file (SVG format) - will be overwritten
  -timestamp=0: Visualize dataset identified by timestamp - leave out or 0 to use latest
```
```
$ maping --help
Usage of maping:
  -config="config.json": Path to configuration file
```

#Example output
Please note this is a rasterized representation of the original vector-based SVG output of the gensvg command
The values are given in seconds between sending a test mail and the last SMTP server in the chain receiving the mail. This assumes every tested host has its time synced, e.g. via NTP.

![maping_svg](https://cloud.githubusercontent.com/assets/6495713/6999310/8eca4062-dc05-11e4-9e15-9bdca1cf676a.png)

#License
Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
Use of this source code is governed by the Apache License v2.0
which can be found in the LICENSE file.
