// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

// config.go
package main

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
const config_file_relative string = "config.json"
const config_file_absolute string = "/etc/maping.json"

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

	database_settings := &databaseSettings{database["file"].(string), database["inmemory"].(bool)}
	//combination of float64 assertion and int64 conversion needed for JSON
	imap_settings := &imapSettings{int(imap["loadrecent"].(float64)), int(imap["timeout"].(float64)), int64(imap["timeoutrcv"].(float64)), int(imap["waittime"].(float64))}

	settings := &configurationData{true, testBody, workerRoutines, mailAccounts, *database_settings, *imap_settings}

	return settings, nil

}

func init() {
	const logprefix string = "Configuration: "
	var (
		config_file = flag.String("config", config_file_relative, "Path to configuration file")
		objmap      map[string]*json.RawMessage
		err         error
	)
	flag.Parse()

	if _, err := os.Stat(*config_file); err == nil {
		objmap, err = readConfig(*config_file)
		if err != nil {
			log.Fatal(logprefix + err.Error())
		}
	} else {
		log.Println(*config_file + " not found. Trying defaults " + config_file_relative + " and " + config_file_absolute)

		if _, err := os.Stat(config_file_relative); err == nil {
			log.Println("Using " + config_file_relative)
			objmap, err = readConfig(config_file_relative)
			if err != nil {
				log.Fatal(logprefix + err.Error())
			}
		} else if _, err := os.Stat(config_file_absolute); err == nil {
			log.Println("Using " + config_file_absolute)
			objmap, err = readConfig(config_file_absolute)
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
