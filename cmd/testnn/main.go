package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cfstras/lfmnn/ffnn"
)

func main() {
	fmt.Println("hai")
	nn := ffnn.New(3, 1, 3, 2)
	defer save(&nn)

	inp := []float32{1, 1, 1}
	out := nn.Update(inp)
	fmt.Println(inp)
	fmt.Println(out)
}

func save(nn *ffnn.NN) {
	if b, err := json.MarshalIndent(&nn, "", "  "); err != nil {
		fmt.Println(err)
		return
	} else {
		if err := ioutil.WriteFile("test-nn.json", b, 0644); err != nil {
			fmt.Println(err)
			return
		}
	}
}
