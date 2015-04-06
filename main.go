// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

//For the scope of this project TX/RX is defined as follows:
//TX: Acc A -> Acc B (using Acc A SMTP and Acc B IMAP)
//RX: Acc B -> Acc A (using Acc B SMTP and Acc A IMAP)

//Since we are using ql as database, you need to build with tag "purego"
//if you want to avoid any CGO dependency while building
//This disables some speed optimizations for the built-in database ql

package main

import (
	"database/sql"
	"log"
	"os"
	"time"
)

type EmailAccount struct {
	Username, Password, SMTPServer, IMAPServer string
	ExplicitSSL_IMAP, ExplicitSSL_SMTP         bool
}

type EmailTemplateData struct {
	From, To, Subject, Body string
}

type EmailAccounts struct {
	acc_a, acc_b *EmailAccount
}

type WorkerResult struct {
	res  *Result
	accs *EmailAccounts
}

func pingWorker(id int, timestamp int64, workerJobs <-chan EmailAccounts, workerOutput chan<- WorkerResult) {
	for j := range workerJobs {
		log.Printf("Worker %v: Testing job %v <-> %v", id, j.acc_a.Username, j.acc_b.Username)
		result := ping(id, j.acc_a, j.acc_b)

		workerOutput <- *&WorkerResult{result, &EmailAccounts{j.acc_a, j.acc_b}}
	}
}

func main() {

	log.SetOutput(os.Stdout)

	var (
		db *sql.DB
		//Timestamp execution to provide fixed data sets
		timestamp    int64              = time.Now().Unix()
		wr           int                = config.workerRoutines
		workerJobs   chan EmailAccounts = make(chan EmailAccounts, 50)
		workerOutput chan WorkerResult  = make(chan WorkerResult, 50)
	)

	db, err := openAndInitDatabase(config.databaseSettings.inmemory, config.databaseSettings.file)

	if err != nil {
		log.Fatal("Error opening/initializing database")
	}

	defer db.Close()

	t := generateTestPairs(len(config.mailAccounts))

	if wr < 1 {
		//Start at least one worker to process testing
		log.Printf("At least 1 worker must be set. Starting 1 instead of %v.", wr)
		wr = 1
	}

	for p := 1; p <= wr; p++ {
		go pingWorker(p, timestamp, workerJobs, workerOutput)
	}

	for _, acc := range t {

		acc_a := acc[0]
		acc_b := acc[1]

		emailacc_a := &EmailAccount{config.mailAccounts[acc_a]["user"].(string), config.mailAccounts[acc_a]["password"].(string), config.mailAccounts[acc_a]["smtphost"].(string), config.mailAccounts[acc_a]["imaphost"].(string), config.mailAccounts[acc_a]["explicitssl_imap"].(bool), config.mailAccounts[acc_a]["explicitssl_smtp"].(bool)}
		emailacc_b := &EmailAccount{config.mailAccounts[acc_b]["user"].(string), config.mailAccounts[acc_b]["password"].(string), config.mailAccounts[acc_b]["smtphost"].(string), config.mailAccounts[acc_b]["imaphost"].(string), config.mailAccounts[acc_b]["explicitssl_imap"].(bool), config.mailAccounts[acc_b]["explicitssl_smtp"].(bool)}
		workerJobs <- *&EmailAccounts{emailacc_a, emailacc_b}
	}

	//No more test pairs
	close(workerJobs)

	for i := 1; i <= len(t); i++ {
		r := <-workerOutput
		err := saveData(db, timestamp, r.res, r.accs.acc_a, r.accs.acc_b)
		if err != nil {
			log.Fatal("Error saving dataset", err.Error())
		}
	}

}
