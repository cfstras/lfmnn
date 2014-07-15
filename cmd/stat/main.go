package main

import "github.com/cfstras/lfmnn/config"
import "github.com/cfstras/lfmnn/bars"

func main() {
	loader := config.Loader
	//loader.Auth() // not needed
	loader.LoadState()

	tags := loader.TagsFloat()

	bars.Graph(tags[:30])
}
