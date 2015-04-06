#Getting started

You will need to have a working go environment installed. Refer to your distributions manual or support forums on accomplish this.

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
