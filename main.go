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

//mail account ping (maping) - utility for checking sets of mail servers (SMTP/IMAPv4).
//Saves results to database and may generate an SVG data visualization matrix from the results.
//For the moment, please refer to the documentation on https://github.com/nfdesign/maping
package main

//For the scope of this project TX/RX is defined as follows:
//TX: Acc A -> Acc B (using Acc A SMTP and Acc B IMAP)
//RX: Acc B -> Acc A (using Acc B SMTP and Acc A IMAP)

import (
	"database/sql"
	"log"
	"os"
	"time"
)

type emailAccount struct {
	username, password, smtpServer, imapServer string
	explicitSSLIMAP, explicitSSLSMTP           bool
}

type emailAccounts struct {
	accA, accB *emailAccount
}

type workerResult struct {
	res  *result
	accs *emailAccounts
}

func pingWorker(id int, timestamp int64, workerJobs <-chan emailAccounts, workerOutput chan<- workerResult) {
	for j := range workerJobs {
		log.Printf("Worker %v: Testing job %v <-> %v", id, j.accA.username, j.accB.username)
		result := ping(id, j.accA, j.accB)

		workerOutput <- *&workerResult{result, &emailAccounts{j.accA, j.accB}}
	}
}

func main() {

	log.SetOutput(os.Stdout)

	var (
		db *sql.DB
		//Timestamp execution to provide fixed data sets
		timestamp    = time.Now().Unix()
		wr           = config.workerRoutines
		workerJobs   = make(chan emailAccounts, 50)
		workerOutput = make(chan workerResult, 50)
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

		accA := acc[0]
		accB := acc[1]

		emailaccA := &emailAccount{config.mailAccounts[accA]["user"].(string), config.mailAccounts[accA]["password"].(string), config.mailAccounts[accA]["smtphost"].(string), config.mailAccounts[accA]["imaphost"].(string), config.mailAccounts[accA]["explicitssl_imap"].(bool), config.mailAccounts[accA]["explicitssl_smtp"].(bool)}
		emailaccB := &emailAccount{config.mailAccounts[accB]["user"].(string), config.mailAccounts[accB]["password"].(string), config.mailAccounts[accB]["smtphost"].(string), config.mailAccounts[accB]["imaphost"].(string), config.mailAccounts[accB]["explicitssl_imap"].(bool), config.mailAccounts[accB]["explicitssl_smtp"].(bool)}
		workerJobs <- *&emailAccounts{emailaccA, emailaccB}
	}

	//No more test pairs
	close(workerJobs)

	for i := 1; i <= len(t); i++ {
		r := <-workerOutput
		err := saveData(db, timestamp, r.res, r.accs.accA, r.accs.accB)
		if err != nil {
			log.Fatal("Error saving dataset", err.Error())
		}
	}

}
