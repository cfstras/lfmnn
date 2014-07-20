package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"

	"github.com/cfstras/lfmnn/ffnn"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
)

func main() {
	//graph()
	nnImageTest()
}

func nnImageTest() {
	// train a FFNN to output an image given x,y
	nn := load()
	if nn == nil {
		nn = ffnn.New(2, 0, 3, 4)
	}
	defer save(nn)

	w, h := 128, 128

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	inp := []float32{0, 0}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			inp[0], inp[1] = float32(x)/float32(w)*8-4, float32(y)/float32(w)*8-4
			out := nn.Update(inp)
			img.Set(x, y, color.RGBA{
				uint8(out[0] * 255), uint8(out[1] * 255), uint8(out[2] * 255),
				255})
		}
	}

	buf := bytes.NewBuffer(nil)
	png.Encode(buf, img)
	ioutil.WriteFile("image.png", buf.Bytes(), 0644)
}

func graph() {
	// draw three graphs
	nn := load()
	if nn == nil {
		nn = ffnn.New(1, 1, 3, 3)
	}
	defer save(nn)

	plot, err := plot.New()
	if err != nil {
		panic(err)
	}
	plot.Title.Text = "Graph NN"
	plot.X.Label.Text = "x"
	plot.Y.Label.Text = "y"

	w := 20

	xys := make([]plotter.XYs, 3)
	for i := range xys {
		xys[i] = make(plotter.XYs, w)
	}

	for x := 0; x < w; x++ {
		out := nn.Update([]float32{float32(x)*4 - 20})
		for i := range xys {
			xys[i][x].X = float64(x)
			xys[i][x].Y = float64(out[i])
		}
	}
	for i := range xys {
		plotutil.AddLinePoints(plot, xys[i])
	}

	if err := plot.Save(4, 4, "image.png"); err != nil {
		panic(err)
	}
}

func load() *ffnn.NN {
	if b, err := ioutil.ReadFile("test-nn.json"); err != nil {
		fmt.Println(err)
		return nil
	} else {
		var ret ffnn.NN
		err = json.Unmarshal(b, &ret)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return &ret
	}
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
