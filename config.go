package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var configFileLocation = ""

type gedConfig struct {
	DefaultCommit string `json:"defaultCommit"`
}

func getConfig() gedConfig {
	checkConfigFileExists()
	return readConfg()
}

func readConfg() gedConfig {
	var userConfig gedConfig
	configBytes, err := os.ReadFile(configFileLocation)
	if err != nil {
        fmt.Printf("WARNING: Unable to read config file...\n")
        userConfig.DefaultCommit = "origin/main"
        return userConfig
	}

	err = json.NewDecoder(bytes.NewBuffer(configBytes)).Decode(&userConfig)

    if err != nil {
        fmt.Printf("WARNING: Unable to read json config...\n")
        userConfig.DefaultCommit = "origin/main"
        return userConfig
    }

	return userConfig
}

func updateConfigFile(newConfig gedConfig) {
	jsonData, err := json.MarshalIndent(newConfig, "", "    ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(configFileLocation, jsonData, 0600)
}

// checks if the config file exists and creates a default if it does not
func checkConfigFileExists() {
	configDir, err := stringGetConfigDir()
	if err != nil {
        fmt.Printf("WARNING: Error getting config dir...\n")
        return
	}

	configFileLocation = filepath.Join(configDir, "gedConfig.json")

	fileExist := fileExists(configFileLocation)

	// if file does not exist create
	if !fileExist {
		os.Mkdir(configDir, 0700)
		configFile, err := os.Create(configFileLocation)
		if err != nil {
            fmt.Printf("WARNING: Unable to create config file. Aborting...\n")
            return
		}

		defer func() {
			configFile.Close()
		}()

		defaultConfig := &gedConfig{DefaultCommit: "origin/main"}

		jsonData, err := json.MarshalIndent(defaultConfig, "", "    ")

		if err != nil {
            fmt.Printf("WARNING: Unable to create json data...\n")
            return
		}

        fmt.Printf("Creating new config file: %s\n", configFileLocation)
		fmt.Printf("Default Config: %s \n", jsonData)
		_, err = configFile.Write(jsonData)
		if err != nil {
            fmt.Printf("WARNING: Unable to write to config file...\n")
		}

	}
}

func stringGetConfigDir() (string, error) {
	userDir := os.Getenv("USERPROFILE")

	if userDir == "" {
		return "", errors.New("Unable to find config dir")
	}

	path := filepath.Join(userDir, "AppData", "Local", "ged")

	return path, nil
}

func fileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}
