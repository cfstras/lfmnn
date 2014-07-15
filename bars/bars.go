package bars

import (
	"math"
	"os"

	tb "github.com/nsf/termbox-go"
)

type NamedValueSlice interface {
	Name(i int) string
	Value(i int) float32
	Len() int
}

func Graph(data NamedValueSlice) {
	tb.Init()
	defer tb.Close()
	width, _ := tb.Size()

	var min, max float32
	min, max = math.MaxFloat32, -math.MaxFloat32
	maxLen := 0
	for i := 0; i < data.Len(); i++ {
		n, v := data.Name(i), data.Value(i)
		if v > max {
			max = v
		} else if v < min {
			min = v
		}
		if len(n) > maxLen {
			maxLen = len(n)
		}
	}

	lineStart := maxLen + 2

	for i := 0; i < width; i++ {
		tb.SetCell(i, 0, '─', tb.ColorWhite, tb.ColorBlack)
	}
	for i := 0; i < data.Len(); i++ {
		n, v := data.Name(i), data.Value(i)

		line := i + 1
		for p, c := range n {
			tb.SetCell(p+1, line, c, tb.ColorCyan, tb.ColorBlack)
		}

		var lwf float32
		if min < 0 {
			lwf = (v - min) / (max - min)
		} else {
			lwf = v / max
		}
		lw := int(lwf * float32(width-lineStart-1))
		for j := 0; j < lw; j++ {
			tb.SetCell(lineStart+j, line, '█', tb.ColorWhite, tb.ColorBlack)
		}
		//fmt.Println(width, lineStart, v, lw, lwf, min, max)
	}
	for i := 0; i < width; i++ {
		tb.SetCell(i, data.Len()+1, '─', tb.ColorWhite, tb.ColorBlack)
	}

	tb.Flush()

	a := []byte{1}
	os.Stdin.Read(a)
}
