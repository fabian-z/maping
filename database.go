// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package main

//Database definition and helper functions

import (
	"database/sql"
	_ "github.com/cznic/ql/driver"
	"log"
)

const structure = `
	CREATE TABLE IF NOT EXISTS log (
		id  			int64,
		timestamp		int64,
		tx				int64,
		rx				int64,
		acc_a			string,
		acc_b			string,
		smtphost_tx		string,
		imaphost_tx		string,
		smtphost_rx		string,
		imaphost_rx		string
	);
	CREATE TABLE IF NOT EXISTS protocol (
		id  int64,
		smtplog_tx		blob,
		imaplog_tx		blob,
		smtplog_rx		blob,
		imaplog_rx		blob
	);
	`

func saveData(db *sql.DB, timestamp int64, res *result, emailacc_a *emailAccount, emailacc_b *emailAccount) error {

	id, err := getNextID(db, "log")

	if err != nil {
		return err
	}

	err = execInPreparedTransaction(db, "INSERT INTO log VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);",
		id, timestamp, res.tx, res.rx, emailacc_a.username, emailacc_b.username, emailacc_a.smtpServer, emailacc_b.imapServer, emailacc_b.smtpServer, emailacc_a.imapServer)

	if err != nil {
		return err
	}

	err = execInPreparedTransaction(db, "INSERT INTO protocol VALUES ($1, $2, $3, $4, $5);",
		id, GzipByteSlice(res.sl_tx), GzipByteSlice(res.il_tx), GzipByteSlice(res.sl_rx), GzipByteSlice(res.il_rx))

	if err != nil {
		return err
	}
	return nil
}

//Execute query in a new transaction on db as a prepared query with input values args
//Rollback if errors occured, Commit if everything went fine
func execInPreparedTransaction(db *sql.DB, query string, args ...interface{}) error {
	ctx, err := db.Begin()

	stmt, err := ctx.Prepare(query)

	_, err = stmt.Exec(args...)

	if err != nil {
		ctx.Rollback()
		log.Fatal(err)
		return err
	}
	if err = ctx.Commit(); err != nil {
		log.Fatal(err)
		return err
	}
	stmt.Close()
	return nil
}

//Function determining the next available id in specific table
//Needed because AUTO_INCREMENT is missing in ql
func getNextID(db *sql.DB, tablename string) (int64, error) {

	var (
		previd int64
		id     int64
	)

	row := db.QueryRow("SELECT id FROM " + tablename + " ORDER BY id DESC LIMIT 1")
	err := row.Scan(&previd)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			// so we use id = 1
			id = 1
		} else {
			log.Fatal(err)
			return -1, err
		}
	} else {
		id = previd + 1
	}

	return id, nil
}

//Opens database dbfile on disk or in memory,
//setting up structure if not yet done
func openAndInitDatabase(ismemory bool, dbfile string) (*sql.DB, error) {

	var (
		db  *sql.DB
		err error
	)

	if ismemory {
		// RAM DB
		db, err = sql.Open("ql-mem", dbfile)

	} else {
		// Disk file DB
		db, err = sql.Open("ql", dbfile)

	}

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	ctx, err := db.Begin()
	_, err = ctx.Exec(structure)

	if err != nil {
		ctx.Rollback()
		log.Fatal(err)

		return nil, err
	}
	err = ctx.Commit()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}
