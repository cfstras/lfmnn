package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cfstras/lfmnn/load"
)

type Main struct {
	loader *load.Loader
	config map[string]string
}

func main() {
	var m Main

	m.LoadConfig()
	defer m.SaveConfig()

	if m.config["username"] == "" {
		fmt.Println("Please set a username in config.json")
		return
	}
	m.loader = load.NewLoader(m.config["username"], m.config["apikey"], m.config["secret"])
	//m.loader.Auth() // not needed
	m.loader.Load()
	defer m.loader.Save()

	m.loader.LoadTracks()
}

func (m *Main) LoadConfig() {
	// defaults
	defaults := map[string]string{
		"apikey":   "",
		"secret":   "",
		"username": "",
	}
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		m.config = defaults
		return
	}
	b, err := ioutil.ReadFile("config.json")
	p(err)
	err = json.Unmarshal(b, &m.config)
	p(err)

	for k, v := range defaults {
		if _, ok := m.config[k]; !ok {
			m.config[k] = v
		}
	}
}

func (m *Main) SaveConfig() {
	b, err := json.MarshalIndent(m.config, "", "  ")
	p(err)
	p(ioutil.WriteFile("config.json", b, 0666))
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}
