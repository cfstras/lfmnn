package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cfstras/lfmnn/load"
)

var (
	Loader *load.Loader
	config map[string]string
)

func init() {
	loadConfig()
	defer saveConfig()

	if config["username"] == "" {
		panic("Please set a username in config.json")
	}
	Loader = load.NewLoader(config["username"],
		config["apikey"], config["secret"])
	//m.loader.Auth() // not needed
}

func loadConfig() {
	// defaults
	defaults := map[string]string{
		"apikey":   "",
		"secret":   "",
		"username": "",
	}
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		config = defaults
		return
	}
	b, err := ioutil.ReadFile("config.json")
	p(err)
	err = json.Unmarshal(b, &config)
	p(err)

	for k, v := range defaults {
		if _, ok := config[k]; !ok {
			config[k] = v
		}
	}
}

func saveConfig() {
	b, err := json.MarshalIndent(config, "", "  ")
	p(err)
	p(ioutil.WriteFile("config.json", b, 0666))
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}
