// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"github.com/ajstarks/svgo"
	_ "github.com/cznic/ql/driver"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//Opens database dbfile on disk
func openAndInitDatabase(dbfile string) (*sql.DB, error) {

	var (
		db  *sql.DB
		err error
	)

	db, err = sql.Open("ql", dbfile)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

func getLastTimestamp(db *sql.DB, tablename string) (int64, error) {

	var laststamp int64

	row := db.QueryRow("SELECT timestamp FROM " + tablename + " ORDER BY timestamp DESC LIMIT 1")
	err := row.Scan(&laststamp)

	if err != nil {
		log.Fatal(err)
		return -1, err
	}

	return laststamp, nil
}

type logentry struct {
	tx, rx                                 int64
	acc_a, acc_b, imaphost_tx, imaphost_rx string
}

const gridwidth int = 50

const fontstyle string = "font-size:12pt;font-family:'source_sans_pro';color:black;"
const textstyle_middle string = fontstyle + "text-anchor:middle; "
const textstyle_left string = fontstyle + "text-anchor:left; "

const red string = "#FF0000"
const yellow string = "#FFFF00"
const green string = "#00FF00"
const grey string = "#ABABAB"
const stroke string = "stroke:black;"

func main() {

	var (
		db            *sql.DB
		canvas        *svg.SVG
		err           error
		laststamp     int64
		height, width int
		stamptime     time.Time
		logarray      []logentry
		cord          map[string]int = make(map[string]int)
		domain        map[int]string = make(map[int]string)
		database                     = flag.String("db", "maping.db", "Path to database")
		output                       = flag.String("output", "output.svg", "Output file (SVG format) - will be overwritten")
		timestamp                    = flag.Int64("timestamp", 0, "Visualize dataset identified by timestamp - leave out or 0 to use latest")
		s             string
		hoster_a      string
		hoster_b      string
		svg_text      string
	)
	flag.Parse()

	if _, err := os.Stat(*database); err == nil {

		db, err = openAndInitDatabase(*database)
	} else {
		log.Println(*database + " not found. Trying defaults maping.db and ../maping.db")

		if _, err := os.Stat("maping.db"); err == nil {
			log.Println("Using maping.db")
			db, err = openAndInitDatabase("maping.db")
		} else if _, err := os.Stat("../maping.db"); err == nil {
			log.Println("Using ../maping.db")
			db, err = openAndInitDatabase("../maping.db")
		} else {
			log.Fatalln("Failed to find database. Giving up..")
		}

	}

	if err != nil {
		log.Fatal(err.Error())
	}

	if *timestamp == 0 || *timestamp < 0 || *timestamp > time.Now().Unix() {

		laststamp, err = getLastTimestamp(db, "log")

		if err != nil {
			log.Fatal(err.Error())
		}

	} else {
		laststamp = *timestamp

	}

	stamptime = time.Unix(laststamp, 0)

	rows, err := db.Query("SELECT tx, rx, acc_a, acc_b, imaphost_tx, imaphost_rx FROM log WHERE timestamp == $1", laststamp)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {

		l := new(logentry)
		err = rows.Scan(&l.tx, &l.rx, &l.acc_a, &l.acc_b, &l.imaphost_tx, &l.imaphost_rx)
		if err != nil {
			log.Fatal(err)

		}

		logarray = append(logarray, *l)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	rows.Close()
	db.Close()

	log.Println("Writing to " + *output)

	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}

	canvas = svg.New(file)

	for _, value := range logarray {
		if len(cord) == 0 {
			cord[value.acc_a] = 1
			cord[value.acc_b] = 2
		} else {
			if cord[value.acc_a] == 0 {
				cord[value.acc_a] = len(cord) + 1
			}
			if cord[value.acc_b] == 0 {
				cord[value.acc_b] = len(cord) + 1
			}
		}

		if hb := strings.Split(value.imaphost_tx, "."); len(hb) > 2 {
			hoster_b = hb[1] + "." + hb[2]
		} else {
			hoster_b = value.imaphost_tx
		}
		if ha := strings.Split(value.imaphost_rx, "."); len(ha) > 2 {
			hoster_a = ha[1] + "." + ha[2]
		} else {
			hoster_a = value.imaphost_rx
		}

		domain[cord[value.acc_a]] = strings.Split(value.acc_a, "@")[1] + " (" + hoster_a + ")"
		domain[cord[value.acc_b]] = strings.Split(value.acc_b, "@")[1] + " (" + hoster_b + ")"
	}

	var amount int = len(domain)

	width = 190 + amount*50
	height = 130 + amount*80

	canvas.Start(width, height)

	//This embeds Source Sans Pro as WOFF into the SVG
	//to achieve a consistent look across platforms
	//See http://caniuse.com/woff and the font def.
	//in font.go
	_, err = canvas.Writer.Write(font_sourcesans)

	if err != nil {
		log.Fatal(err)
	}

	canvas.Grid(100, 100, amount*gridwidth, amount*gridwidth, gridwidth, stroke+"stroke-width:0.25pt;")

	canvas.Line(90, 70, 60, 100, stroke)

	canvas.Line(90, 70, 70, 80, stroke)
	canvas.Line(90, 70, 80, 90, stroke)
	canvas.Text(170, 40, stamptime.Format("2 Jan 06, 15:04 (MST)"), textstyle_middle)

	for i := amount - 1; i >= 0; i-- {
		pos := (i * gridwidth) + 100
		canvas.Rect(pos, pos, gridwidth, gridwidth, "fill:"+grey)

		canvas.Text(70, pos+30, strconv.Itoa(i+1), textstyle_middle)
		canvas.Text(pos+20, 80, strconv.Itoa(i+1), textstyle_middle)

	}

	for _, value := range logarray {

		var x, y int
		var color_tx, color_rx string
		x = 50 + cord[value.acc_b]*50
		y = 50 + cord[value.acc_a]*50

		text_tx := strconv.FormatInt(value.tx, 10)
		text_rx := strconv.FormatInt(value.rx, 10)

		switch {
		case value.tx < 0:
			color_tx = red
			text_tx = "ERR"
		case value.tx == 0:
			color_tx = green
			text_tx = "<1"
		case value.tx <= 30:
			color_tx = green
		case value.tx > 30:
			color_tx = yellow
		}

		switch {
		case value.rx < 0:
			color_rx = red
			text_rx = "ERR"
		case value.rx == 0:
			color_rx = green
			text_rx = "<1"
		case value.rx <= 30:
			color_rx = green
		case value.rx > 30:
			color_rx = yellow
		}

		canvas.Rect(x, y, gridwidth, gridwidth, "fill:"+color_tx)
		canvas.Text(x+22, y+30, text_tx, textstyle_middle)

		canvas.Rect(y, x, gridwidth, gridwidth, "fill:"+color_rx)
		canvas.Text(y+22, x+30, text_rx, textstyle_middle)

	}

	for i := 1; i <= len(domain); i++ {

		s = s + `<tspan x="100" dy="15">` + strconv.Itoa(i) + ": " + domain[i] + "</tspan>"

	}

	svg_text = `<text x="0" y="` + strconv.Itoa(width-40) + `" style="` + textstyle_left + `">` + s + "</text>"
	_, err = canvas.Writer.Write([]byte(svg_text))
	if err != nil {
		log.Fatal(err)
	}
	canvas.End()
	file.Close()
}
