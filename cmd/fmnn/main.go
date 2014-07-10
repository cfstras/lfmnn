package main

import (
	"fmt"

	"github.com/cfstras/lfmnn/config"
	"github.com/cfstras/lfmnn/ffnn"
)

func main() {
	loader := config.Loader
	loader.LoadState()
	defer loader.SaveState()
	loader.LoadTracksAndTags()

	nn := ffnn.New(3, 1, 3, 2)

	inp := []float32{1, 1, 1}
	out := nn.Update(inp)
	fmt.Println(inp)
	fmt.Println(out)
}
