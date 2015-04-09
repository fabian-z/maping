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

package imapclient

import (
	"testing"
	"time"
)

const utilsTestLogPrefix = "date_test.go: "

func TestDateParse(t *testing.T) {

	var (
		testdate = "Mon, 02 Jan 2005 15:04:06 -0700 (MST)"
		timet    time.Time
		err      error
	)

	timet, err = parseDate(testdate)
	if err != nil {
		t.Fatal("Error parsing date: " + err.Error())
	}

	if timet.Format(time.RFC822) != "02 Jan 05 15:04 MST" {
		t.Fatal("Error parsing date, got: " + timet.Format(time.RFC822))
	}

	testdate = "2 Jan 05 15:04 CET"

	timet, err = parseDate(testdate)
	if err != nil {
		t.Fatal("Error parsing date: " + err.Error())
	}

	if timet.Format(time.RFC822) != "02 Jan 05 15:04 CET" {
		t.Fatal("Error parsing date, got: " + timet.Format(time.RFC822))
	}

}
