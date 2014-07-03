package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/cfstras/lfmnn/load"
)

type Main struct {
	loader load.Loader
	config map[string]string
}

func main() {
	var m Main
	m.LoadConfig()
	defer m.SaveConfig()
	m.loader = load.NewLoader(m.config["apikey"], m.config["secret"])

}

func (m *Main) LoadConfig() {
	// defaults
	defaults := map[string]string{
		"apikey":        "",
		"secret":        "",
		"earliestTrack": fmt.Sprint(time.Now().Unix()),
		"numTracks":     "0",
	}
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		m.config = defaults
		return
	}
	f, err := os.Open("config.json")
	p(err)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
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
	f, err := os.Create("config.json")
	p(err)
	defer f.Close()

	b, err := json.MarshalIndent(m.config, "", "  ")
	p(err)
	_, err = f.Write(b)
	p(err)
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}
