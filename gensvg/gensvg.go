//Copyright 2015 NF Design UG (haftungsbeschraenkt)
//All right reserved.

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at

//  http://www.apache.org/licenses/LICENSE-2.0

//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

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
	tx, rx                             int64
	accA, accB, imapHostTx, imapHostRx string
}

const gridwidth int = 50

const fontStyle string = "font-size:12pt;font-family:'source_sans_pro';color:black;"
const textStyleMiddle string = fontStyle + "text-anchor:middle; "
const textStyleLeft string = fontStyle + "text-anchor:left; "

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
		cord          = make(map[string]int)
		domain        = make(map[int]string)
		database      = flag.String("db", "maping.db", "Path to database")
		output        = flag.String("output", "output.svg", "Output file (SVG format) - will be overwritten")
		timestamp     = flag.Int64("timestamp", 0, "Visualize dataset identified by timestamp - leave out or 0 to use latest")
		s             string
		hosterA       string
		hosterB       string
		svgText       string
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
		err = rows.Scan(&l.tx, &l.rx, &l.accA, &l.accB, &l.imapHostTx, &l.imapHostRx)
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
			cord[value.accA] = 1
			cord[value.accB] = 2
		} else {
			if cord[value.accA] == 0 {
				cord[value.accA] = len(cord) + 1
			}
			if cord[value.accB] == 0 {
				cord[value.accB] = len(cord) + 1
			}
		}

		if hb := strings.Split(value.imapHostTx, "."); len(hb) > 2 {
			hosterB = hb[1] + "." + hb[2]
		} else {
			hosterB = value.imapHostTx
		}
		if ha := strings.Split(value.imapHostRx, "."); len(ha) > 2 {
			hosterA = ha[1] + "." + ha[2]
		} else {
			hosterA = value.imapHostRx
		}

		domain[cord[value.accA]] = strings.Split(value.accA, "@")[1] + " (" + hosterA + ")"
		domain[cord[value.accB]] = strings.Split(value.accB, "@")[1] + " (" + hosterB + ")"
	}

	var amount = len(domain)

	width = 190 + amount*50
	height = 130 + amount*80

	canvas.Start(width, height)

	//This embeds Source Sans Pro as WOFF into the SVG
	//to achieve a consistent look across platforms
	//See http://caniuse.com/woff and the font def.
	//in font.go
	_, err = canvas.Writer.Write(fontSourceSansPro)

	if err != nil {
		log.Fatal(err)
	}

	canvas.Grid(100, 100, amount*gridwidth, amount*gridwidth, gridwidth, stroke+"stroke-width:0.25pt;")

	canvas.Line(90, 70, 60, 100, stroke)

	canvas.Line(90, 70, 70, 80, stroke)
	canvas.Line(90, 70, 80, 90, stroke)
	canvas.Text(170, 40, stamptime.Format("2 Jan 06, 15:04 (MST)"), textStyleMiddle)

	for i := amount - 1; i >= 0; i-- {
		pos := (i * gridwidth) + 100
		canvas.Rect(pos, pos, gridwidth, gridwidth, "fill:"+grey)

		canvas.Text(70, pos+30, strconv.Itoa(i+1), textStyleMiddle)
		canvas.Text(pos+20, 80, strconv.Itoa(i+1), textStyleMiddle)

	}

	for _, value := range logarray {

		var x, y int
		var colorTx, colorRx string
		x = 50 + cord[value.accB]*50
		y = 50 + cord[value.accA]*50

		textTx := strconv.FormatInt(value.tx, 10)
		textRx := strconv.FormatInt(value.rx, 10)

		switch {
		case value.tx < 0:
			colorTx = red
			textTx = "ERR"
		case value.tx == 0:
			colorTx = green
			textTx = "<1"
		case value.tx <= 30:
			colorTx = green
		case value.tx > 30:
			colorTx = yellow
		}

		switch {
		case value.rx < 0:
			colorRx = red
			textRx = "ERR"
		case value.rx == 0:
			colorRx = green
			textRx = "<1"
		case value.rx <= 30:
			colorRx = green
		case value.rx > 30:
			colorRx = yellow
		}

		canvas.Rect(x, y, gridwidth, gridwidth, "fill:"+colorTx)
		canvas.Text(x+22, y+30, textTx, textStyleMiddle)

		canvas.Rect(y, x, gridwidth, gridwidth, "fill:"+colorRx)
		canvas.Text(y+22, x+30, textRx, textStyleMiddle)

	}

	for i := 1; i <= len(domain); i++ {

		s = s + `<tspan x="100" dy="15">` + strconv.Itoa(i) + ": " + domain[i] + "</tspan>"

	}

	svgText = `<text x="0" y="` + strconv.Itoa(width-40) + `" style="` + textStyleLeft + `">` + s + "</text>"
	_, err = canvas.Writer.Write([]byte(svgText))
	if err != nil {
		log.Fatal(err)
	}
	canvas.End()
	file.Close()
}
