// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package main

//Reading and parsing configuration

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var config *configurationData

type configurationData struct {
	initialized    bool
	testBody       string
	workerRoutines int
	mailAccounts   []map[string]interface{}
	databaseSettings
	imapSettings
}

type databaseSettings struct {
	file     string
	inmemory bool
}

type imapSettings struct {
	loadRecent int
	timeout    int
	timeoutRcv int64
	waitTime   int
}

//To be given with absolute or relative path to default configuration file
const configFileRelative string = "config.json"
const configFileAbsolute string = "/etc/maping.json"

// readConfig reads JSON-style configuration from file.
// It removes comments, unmarshals the top level and puts
// the result into a map[string]*json.RawMessage
func readConfig(configfile string) (map[string]*json.RawMessage, error) {

	var (
		file   []byte
		objmap map[string]*json.RawMessage
		err    error
	)

	file, err = ioutil.ReadFile(configfile)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(file), "\n")
	var jdata []byte
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			jdata = append(jdata, []byte(line+"\n")...)
		}
	}

	err = json.Unmarshal(jdata, &objmap)
	if err != nil {
		return nil, err
	}
	return objmap, nil
}

func parseConfig(objmap map[string]*json.RawMessage) (*configurationData, error) {

	var (
		err            error
		testBody       string
		workerRoutines int
	)

	err = json.Unmarshal(*objmap["testBody"], &testBody)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(*objmap["workerRoutines"], &workerRoutines)
	if err != nil {
		return nil, err
	}

	var mailAccounts []map[string]interface{}
	err = json.Unmarshal(*objmap["mailAccounts"], &mailAccounts)
	if err != nil {
		return nil, err
	}

	var database map[string]interface{}
	err = json.Unmarshal(*objmap["database"], &database)
	if err != nil {
		return nil, err
	}

	var imap map[string]interface{}
	err = json.Unmarshal(*objmap["imap"], &imap)
	if err != nil {
		return nil, err
	}

	dSettings := &databaseSettings{database["file"].(string), database["inmemory"].(bool)}
	//combination of float64 assertion and int64 conversion needed for JSON
	iSettings := &imapSettings{int(imap["loadrecent"].(float64)), int(imap["timeout"].(float64)), int64(imap["timeoutrcv"].(float64)), int(imap["waittime"].(float64))}

	settings := &configurationData{true, testBody, workerRoutines, mailAccounts, *dSettings, *iSettings}

	return settings, nil

}

func init() {
	const logprefix string = "Configuration: "
	var (
		configFile = flag.String("config", configFileRelative, "Path to configuration file")
		objmap     map[string]*json.RawMessage
		err        error
	)
	flag.Parse()

	if _, err := os.Stat(*configFile); err == nil {
		objmap, err = readConfig(*configFile)
		if err != nil {
			log.Fatal(logprefix + err.Error())
		}
	} else {
		log.Println(*configFile + " not found. Trying defaults " + configFileRelative + " and " + configFileAbsolute)

		if _, err := os.Stat(configFileRelative); err == nil {
			log.Println("Using " + configFileRelative)
			objmap, err = readConfig(configFileRelative)
			if err != nil {
				log.Fatal(logprefix + err.Error())
			}
		} else if _, err := os.Stat(configFileAbsolute); err == nil {
			log.Println("Using " + configFileAbsolute)
			objmap, err = readConfig(configFileAbsolute)
			if err != nil {
				log.Fatal(logprefix + err.Error())
			}
		} else {
			log.Fatalln(logprefix + "Failed to find configuration file. Giving up..")
		}

	}

	if err != nil {
		log.Fatal(logprefix + err.Error())
	}

	config, err = parseConfig(objmap)

	if err != nil {
		log.Fatal(logprefix + err.Error())
	}

}
