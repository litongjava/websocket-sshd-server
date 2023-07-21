package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

var CONFIG *Config

func ReadFile(filename string) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err.Error())
	}

	err = yaml.Unmarshal(yamlFile, &CONFIG)
	if err != nil {
		fmt.Println("error", err.Error())
	}
}
