package main

import (
	"fmt"
	"math"
	"math/rand"
)

const goldenRatioConjugate = 0.618033988749895

type ColorGenerator struct {
	hue        float64
	saturation float64
	lightness  float64
}

func NewColorGenerator() *ColorGenerator {
	return &ColorGenerator{
		hue:        rand.Float64(),
		saturation: 0.5,
		lightness:  0.5,
	}
}

func (cg *ColorGenerator) CreateHex() string {
	cg.hue = math.Mod(cg.hue+goldenRatioConjugate, 1.0)
	r, g, b := hslToRgb(cg.hue, cg.saturation, cg.lightness)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func hslToRgb(h, s, l float64) (r, g, b uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h*6, 2)-1))
	m := l - c/2

	var r1, g1, b1 float64

	switch {
	case h < 1.0/6.0:
		r1, g1, b1 = c, x, 0
	case h < 2.0/6.0:
		r1, g1, b1 = x, c, 0
	case h < 3.0/6.0:
		r1, g1, b1 = 0, c, x
	case h < 4.0/6.0:
		r1, g1, b1 = 0, x, c
	case h < 5.0/6.0:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}

	r = uint8((r1 + m) * 255)
	g = uint8((g1 + m) * 255)
	b = uint8((b1 + m) * 255)

	return
}
