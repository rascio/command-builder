package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func logError(str string, params ...any) {
	fmt.Fprintf(os.Stderr, str, params...)
}

func loadSpec(conf *map[string]CmdSpec, name string) {
	found, err := Exists(name)
	if found && err == nil {
		yamlFile, err := ioutil.ReadFile(name)
		if err != nil {
			logError(fmt.Sprintf("Error reading: %s", name), err)
		}
		err = yaml.Unmarshal(yamlFile, conf)
		if err != nil {
			logError(fmt.Sprintf("Error parsing: %s", name), err)
		}
	}
}
