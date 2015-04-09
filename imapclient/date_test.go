// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

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
