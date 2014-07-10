package main

import "github.com/cfstras/lfmnn/config"

func main() {
	loader := config.Loader
	//loader.Auth() // not needed
	loader.LoadState()
	defer loader.SaveState()

	loader.LoadTracksAndTags()
}
