package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
	_ "net/http/pprof"

	"github.com/cfstras/lfmnn/ffnn"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"

	"github.com/thoj/go-galib"
)

func main() {
	go http.ListenAndServe(":6060", nil)

	//graph()
	nnImageTest()
}

func nnImageTest() {
	// train a FFNN to output an image given x,y
	var nn *ffnn.NN //load()
	if nn == nil {
		nn = ffnn.New(2, 2, 3, 16)
	}
	defer save(nn)

	genomeLength := 0
	for _, l := range nn.Layers {
		for _, n := range l {
			genomeLength += len(n.Weights)
		}
	}

	w, h := 128, 128
	rect := image.Rect(0, 0, w, h)

	img := image.NewRGBA(rect)
	inputImage := image.NewRGBA(rect)
	inp := []float32{0, 0}

	fmt.Println("loading input...")
	inpB, err := ioutil.ReadFile("input.png")
	p(err)
	inpBuf := bytes.NewBuffer(inpB)
	img1, err := png.Decode(inpBuf)
	p(err)
	draw.Draw(inputImage, rect, img1, img1.Bounds().Min, draw.Src)

	fmt.Println("ga init...")

	m := ga.NewMultiMutator()
	m.Add(new(ga.GAShiftMutator))
	m.Add(new(ga.GASwitchMutator))
	m.Add(ga.NewGAGaussianMutator(1, 0))

	param := ga.GAParameter{
		Initializer: new(ga.GARandomInitializer),
		Selector:    ga.NewGATournamentSelector(0.2, 20),
		Breeder:     new(ga.GA2PointBreeder),
		Mutator:     m,
		PMutate:     0.3,
		PBreed:      0.4}

	gao := ga.NewGA(param)

	coord := func(img *image.RGBA, x, y int) int {
		return (y-img.Rect.Min.Y)*img.Stride + (x-img.Rect.Min.X)*4
	}

	saveImage := func() {
		name := "image-" + fmt.Sprint(rand.Int()) + ".png"
		fmt.Println("saving", name)
		buf := bytes.NewBuffer(nil)
		png.Encode(buf, img)
		ioutil.WriteFile(name, buf.Bytes(), 0644)
	}

	makeImage := func() {
		fmt.Println("generating image...")
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				inp[0], inp[1] = float32(x)/float32(w)*8-4, float32(y)/float32(w)*8-4
				out := nn.Update(inp)

				i := coord(img, x, y)
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] =
					uint8(out[0]*255), uint8(out[1]*255), uint8(out[2]*255), 255
			}
		}
	}
	setNN := func(g *ga.GAFloatGenome) {
		i := 0
		for li := range nn.Layers {
			for ni := range nn.Layers[li] {
				for vi := range nn.Layers[li][ni].Weights {
					nn.Layers[li][ni].Weights[vi] = float32(g.Gene[i])
					i++
				}
			}
		}
	}

	scores := 0
	contrast := func(g *ga.GAFloatGenome) float64 {
		// set values
		setNN(g)

		makeImage()

		// calc contrast to input
		var contrast float64
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				i := coord(img, x, y)
				a_r, a_g, a_b := inputImage.Pix[i], inputImage.Pix[i+1], inputImage.Pix[i+2]
				b_r, b_g, b_b := img.Pix[i], img.Pix[i+1], img.Pix[i+2]
				dr, dg, db := float64(a_r)-float64(b_r),
					float64(a_g)-float64(b_g), float64(a_b)-float64(b_b)
				if dr < 0 {
					dr *= -1
				}
				if dg < 0 {
					dg *= -1
				}
				if db < 0 {
					db *= -1
				}
				contrast += dr + dg + db
			}
		}
		fmt.Println("contrast:", contrast)
		scores++
		if scores%50 == 0 {
			saveImage()
		}
		return contrast
	}

	genome := ga.NewFloatGenome(make([]float64, genomeLength), contrast, -20, 20)

	gao.Init(20, genome)

	fmt.Println("running...")
	gao.Optimize(100)

	fmt.Println("best score:", gao.Best().Score())

	setNN(gao.Best().(*ga.GAFloatGenome))
	makeImage()
	saveImage()
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

func p(err error) {
	if err != nil {
		panic(err)
	}
}
