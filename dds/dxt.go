package dds

import "image/color"

type RGB565 uint16

func (c RGB565) rgb24() (r, g, b uint8) {
	r = uint8((c>>11)&(1<<5-1)) << 3
	r |= r >> 5
	g = uint8((c>>5)&(1<<6-1)) << 2
	g |= g >> 6
	b = uint8(c&(1<<5-1)) << 3
	b |= b >> 5
	return
}

func (c RGB565) RGBA() (r, g, b, a uint32) {
	// first expand to 8 bits, then to 16
	r8, g8, b8 := c.rgb24()
	return color.RGBA{r8, g8, b8, 0xff}.RGBA()
}

func packRGB(r, g, b uint8) RGB565 {
	return RGB565((uint16(r)>>3)<<11 | (uint16(g)>>2)<<5 | uint16(b)>>3)
}

var (
	RGB565Model = color.ModelFunc(rgb565Model)
)

func rgb565Model(c color.Color) color.Color {
	if _, ok := c.(RGB565); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return packRGB(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func decodeDxt1Block(pix []uint8, b []byte, stride int, dxt3 bool) {
	if len(b) < 8 {
		panic("not enough data to decode block")
	}

	c0 := uint16(b[1])<<8 | uint16(b[0])
	c1 := uint16(b[3])<<8 | uint16(b[2])
	codes := uint32(b[7])<<24 | uint32(b[6])<<16 | uint32(b[5])<<8 | uint32(b[4])
	palette := mkPalette(c0, c1, c0 > c1 || dxt3)

	for i := uint(0); i < 16; i++ {
		ii := (i&3)<<2 + (i>>2)*uint(stride)
		c := (codes >> (i << 1)) & 3
		pix[ii+0] = palette[c*3+0]
		pix[ii+1] = palette[c*3+1]
		pix[ii+2] = palette[c*3+2]
		pix[ii+3] = 0xff
	}
}

func mkPalette(c0, c1 uint16, qColor bool) []uint8 {
	r0, g0, b0 := RGB565(c0).rgb24()
	r1, g1, b1 := RGB565(c1).rgb24()
	if qColor {
		r2 := (2*uint16(r0) + uint16(r1)) / 3
		g2 := (2*uint16(g0) + uint16(g1)) / 3
		b2 := (2*uint16(b0) + uint16(b1)) / 3
		r3 := (uint16(r0) + 2*uint16(r1)) / 3
		g3 := (uint16(g0) + 2*uint16(g1)) / 3
		b3 := (uint16(b0) + 2*uint16(b1)) / 3
		return []uint8{
			r0, g0, b0,
			r1, g1, b1,
			uint8(r2), uint8(g2), uint8(b2),
			uint8(r3), uint8(g3), uint8(b3),
		}
	}
	r2 := (uint16(r0) + uint16(r1)) / 2
	g2 := (uint16(g0) + uint16(g1)) / 2
	b2 := (uint16(b0) + uint16(b1)) / 2
	return []uint8{
		r0, g0, b0,
		r1, g1, b1,
		uint8(r2), uint8(g2), uint8(b2),
		0, 0, 0,
	}
}

func decodeDxt1ABlock(pix []uint8, b []byte, stride int) {
	if len(b) < 8 {
		panic("not enough data to decode block")
	}

	c0 := uint16(b[1])<<8 | uint16(b[0])
	c1 := uint16(b[3])<<8 | uint16(b[2])
	codes := uint32(b[7])<<24 | uint32(b[6])<<16 | uint32(b[5])<<8 | uint32(b[4])
	palette := mkPalette(c0, c1, c0 > c1)

	for i := uint(0); i < 16; i++ {
		ii := (i&3)<<2 + (i>>2)*uint(stride)
		c := (codes >> (i << 1)) & 3
		pix[ii+0] = palette[c*3+0]
		pix[ii+1] = palette[c*3+1]
		pix[ii+2] = palette[c*3+2]
		if c0 <= c1 && c == 3 {
			pix[ii+3] = 0
		} else {
			pix[ii+3] = 0xff
		}
	}
}

func decodeDxt3Block(pix []uint8, b []byte, stride int) {
	if len(b) < 16 {
		panic("not enough data to decode block")
	}

	alpha := uint64(b[7])<<56 | uint64(b[6])<<48 | uint64(b[5])<<40 | uint64(b[4])<<32 | uint64(b[3])<<24 | uint64(b[2])<<16 | uint64(b[1])<<8 | uint64(b[0])

	decodeDxt1Block(pix, b[8:], stride, true)
	for i := uint(0); i < 16; i++ {
		ii := (i&3)<<2 + (i>>2)*uint(stride)
		a := (alpha >> (i << 2)) & 0xf
		pix[ii+3] = uint8(a)<<4 | uint8(a)
	}
}

func decodeDxt5Block(pix []uint8, b []byte, stride int) {
	if len(b) < 16 {
		panic("not enough data to decode block")
	}

	a0, a1 := b[0], b[1]
	code := uint64(b[7])<<40 | uint64(b[6])<<32 | uint64(b[5])<<24 | uint64(b[4])<<16 | uint64(b[3])<<8 | uint64(b[2])
	alphaPalette := mkAlphaPalette(a0, a1, a0 > a1)

	decodeDxt1Block(pix, b[8:], stride, true)
	for i := uint(0); i < 16; i++ {
		ii := (i&3)<<2 + (i>>2)*uint(stride)
		c := (code >> (3 * i)) & 7
		pix[ii+3] = alphaPalette[c]
	}
}

func mkAlphaPalette(a0, a1 uint8, upper bool) []uint8 {
	if upper {
		return []uint8{
			a0,
			a1,
			uint8((6*uint16(a0) + 1*uint16(a1)) / 7),
			uint8((5*uint16(a0) + 2*uint16(a1)) / 7),
			uint8((4*uint16(a0) + 3*uint16(a1)) / 7),
			uint8((3*uint16(a0) + 4*uint16(a1)) / 7),
			uint8((2*uint16(a0) + 5*uint16(a1)) / 7),
			uint8((1*uint16(a0) + 6*uint16(a1)) / 7),
		}
	}
	return []uint8{
		a0,
		a1,
		uint8((4*uint16(a0) + 1*uint16(a1)) / 5),
		uint8((3*uint16(a0) + 2*uint16(a1)) / 5),
		uint8((2*uint16(a0) + 3*uint16(a1)) / 5),
		uint8((1*uint16(a0) + 4*uint16(a1)) / 5),
		0,
		255,
	}
}
