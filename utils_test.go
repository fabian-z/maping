// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"testing"
)

const utilsTestLogPrefix = "utils_test.go: "

func TestGetRandomString(t *testing.T) {

	const bytesize int = 15
	var (
		rs  string
		b   []byte
		err error
	)

	rs = GetRandomString(bytesize)
	b, err = base64.StdEncoding.DecodeString(rs)

	if err != nil {
		t.Error(utilsTestLogPrefix + "Error decoding base64 string: " + err.Error())
	}
	if len(b) != bytesize {
		t.Errorf(utilsTestLogPrefix+"Wrong byte size: %v", len(b))
	}

}

func TestGzipByteSlice(t *testing.T) {

	var (
		randomtestdata, compressed []byte
		cbuf, buf                  *bytes.Buffer
		err                        error
		r                          *gzip.Reader
	)

	randomtestdata = []byte(GetRandomString(150))

	compressed = GzipByteSlice(randomtestdata)

	if len(randomtestdata) >= len(compressed) {
		t.Error(utilsTestLogPrefix + "gzip did not properly compress data")
	}

	cbuf = new(bytes.Buffer)
	_, err = cbuf.Write(compressed)
	if err != nil {
		t.Error(err)
	}

	r, err = gzip.NewReader(cbuf)
	if err != nil {
		t.Error(err)
	}
	buf = new(bytes.Buffer)
	buf.ReadFrom(r)
	r.Close()

	if bytes.Compare(buf.Bytes(), randomtestdata) != 0 {
		t.Error(utilsTestLogPrefix + "Returned different data from gzip")
	}

}
