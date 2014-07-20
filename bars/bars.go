package bars

import (
	"fmt"
	"math"

	mu "github.com/cfstras/go-utils/math"
	tb "github.com/nsf/termbox-go"
)

type NamedValueSlice interface {
	Name(i int) string
	Value(i int) float32
	Len() int
}

type IVec2 struct {
	X, Y int
}

type context struct {
	data NamedValueSlice

	minVal, maxVal float32
	res            IVec2
	lineStart      int
	offset         IVec2
	log            bool

	end bool
}

func Graph(data NamedValueSlice) {
	tb.Init()
	defer tb.Close()

	c := &context{data: data}
	c.res.X, c.res.Y = tb.Size()

	c.minVal, c.maxVal = math.MaxFloat32, -math.MaxFloat32
	maxLen := 0
	for i := 0; i < data.Len(); i++ {
		n, v := data.Name(i), data.Value(i)
		if v > c.maxVal {
			c.maxVal = v
		} else if v < c.minVal {
			c.minVal = v
		}
		if len(n) > maxLen {
			maxLen = len(n)
		}
	}

	c.lineStart = mu.MinI(maxLen+2, 15)

	for !c.end {
		c.Draw()

		ev := tb.PollEvent()
		c.Handle(ev)
	}
}
func (c *context) Draw() {

	for i := 0; i < c.res.X; i++ {
		tb.SetCell(i, 0, '─', tb.ColorWhite, tb.ColorBlack)
	}
	minS := fmt.Sprintf("%.4g", c.minVal)
	maxS := fmt.Sprintf("%.4g", c.maxVal)
	for i, v := range maxS {
		tb.SetCell(c.res.X-len(maxS)-1+i, 0, v, tb.ColorWhite, tb.ColorBlack)
	}
	for i, v := range minS {
		tb.SetCell(c.lineStart+i, 0, v, tb.ColorWhite, tb.ColorBlack)
	}

	for i := 0; i < c.data.Len() && i < c.res.Y-2; i++ {
		n, v := c.data.Name(c.offset.Y+i), c.data.Value(c.offset.Y+i)

		line := i + 1
		for p, c := range n {
			tb.SetCell(p+1, line, c, tb.ColorCyan, tb.ColorBlack)
		}

		var lwf float32
		if c.minVal < 0 {
			lwf = (v - c.minVal) / (c.maxVal - c.minVal)
		} else {
			lwf = v / c.maxVal
		}
		lw := int(lwf * float32(c.res.X-c.lineStart-1))
		for j := 0; j < lw; j++ {
			tb.SetCell(c.lineStart+j, line, '█', tb.ColorWhite, tb.ColorBlack)
		}
	}
	for i := 0; i < c.res.X; i++ {
		tb.SetCell(i, c.res.Y-1, '─', tb.ColorWhite, tb.ColorBlack)
	}

	tb.Flush()
	for x := 0; x < c.res.X; x++ {
		for y := 0; y < c.res.Y; y++ {
			tb.SetCell(x, y, ' ', tb.ColorBlack, tb.ColorBlack)
		}
	}
}

func (c *context) Handle(ev tb.Event) {
	switch ev.Type {
	case tb.EventError:
		c.end = true
		fmt.Println(ev.Err)
	case tb.EventKey:
		switch ev.Key {
		case tb.KeyArrowUp:
			c.offset.Y--
		case tb.KeyArrowDown:
			c.offset.Y++
		case tb.KeyPgup:
			c.offset.Y -= c.res.Y
		case tb.KeyPgdn:
			c.offset.Y += c.res.Y
		case tb.KeyArrowLeft:
			c.offset.X++
		case tb.KeyArrowRight:
			c.offset.X--
		case tb.KeyEsc:
			fallthrough
		case tb.KeyEnter:
			c.end = true
		case 'l':
			c.log = !c.log
		}
		c.offset.Y = mu.MinI(c.offset.Y, c.data.Len()-c.res.Y+2)
		c.offset.Y = mu.MaxI(c.offset.Y, 0)
	default:
	}
}
