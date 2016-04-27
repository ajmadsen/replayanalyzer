package dds

import "image/color"

type RGB565 uint16

func (c RGB565) RGBA() (r, g, b, a uint32) {
	// first expand to 8 bits, then to 16
	r8 := uint8((c>>11)&(1<<5-1)) << 3
	r8 |= r8 >> 5
	g8 := uint8((c>>5)&(1<<6-1)) << 2
	g8 |= g8 >> 6
	b8 := uint8(c&(1<<5-1)) << 3
	b8 |= b8 >> 5
	a8 := uint8(0xff)
	return color.RGBA{r8, g8, b8, a8}.RGBA()
}

var (
	RGB565Model = color.ModelFunc(rgb565Model)
)

func rgb565Model(c color.Color) color.Color {
	if _, ok := c.(RGB565); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	nc := (r>>11)<<11 | (g>>10)<<5 | (b >> 11)
	return RGB565(nc)
}
