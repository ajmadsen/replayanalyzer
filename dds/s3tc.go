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

func decodeDtx1Block(b []byte, dxt3 bool) []RGB565 {
	if len(b) < 8 {
		panic("not enough data to decode block")
	}

	c0 := uint16(b[1])<<8 | uint16(b[0])
	c1 := uint16(b[3])<<8 | uint16(b[2])
	if dxt3 && c0 <= c1 {
		c0, c1 = c1, c0
	}
	codes := uint32(b[7])<<24 | uint32(b[6])<<16 | uint32(b[5])<<8 | uint32(b[4])

	return []RGB565{
		cSelect(cMask(codes, 0, 0), c0, c1),
		cSelect(cMask(codes, 1, 0), c0, c1),
		cSelect(cMask(codes, 2, 0), c0, c1),
		cSelect(cMask(codes, 3, 0), c0, c1),
		cSelect(cMask(codes, 0, 1), c0, c1),
		cSelect(cMask(codes, 1, 1), c0, c1),
		cSelect(cMask(codes, 2, 1), c0, c1),
		cSelect(cMask(codes, 3, 1), c0, c1),
		cSelect(cMask(codes, 0, 2), c0, c1),
		cSelect(cMask(codes, 1, 2), c0, c1),
		cSelect(cMask(codes, 2, 2), c0, c1),
		cSelect(cMask(codes, 3, 2), c0, c1),
		cSelect(cMask(codes, 0, 3), c0, c1),
		cSelect(cMask(codes, 1, 3), c0, c1),
		cSelect(cMask(codes, 2, 3), c0, c1),
		cSelect(cMask(codes, 3, 3), c0, c1),
	}
}

func decodeDtx1ABlock(b []byte) ([]RGB565, []uint8) {
	if len(b) < 8 {
		panic("not enough data to decode block")
	}

	c0 := uint16(b[1])<<8 | uint16(b[0])
	c1 := uint16(b[3])<<8 | uint16(b[2])
	codes := uint32(b[7])<<24 | uint32(b[6])<<16 | uint32(b[5])<<8 | uint32(b[4])

	return []RGB565{
			cSelect(cMask(codes, 0, 0), c0, c1),
			cSelect(cMask(codes, 1, 0), c0, c1),
			cSelect(cMask(codes, 2, 0), c0, c1),
			cSelect(cMask(codes, 3, 0), c0, c1),
			cSelect(cMask(codes, 0, 1), c0, c1),
			cSelect(cMask(codes, 1, 1), c0, c1),
			cSelect(cMask(codes, 2, 1), c0, c1),
			cSelect(cMask(codes, 3, 1), c0, c1),
			cSelect(cMask(codes, 0, 2), c0, c1),
			cSelect(cMask(codes, 1, 2), c0, c1),
			cSelect(cMask(codes, 2, 2), c0, c1),
			cSelect(cMask(codes, 3, 2), c0, c1),
			cSelect(cMask(codes, 0, 3), c0, c1),
			cSelect(cMask(codes, 1, 3), c0, c1),
			cSelect(cMask(codes, 2, 3), c0, c1),
			cSelect(cMask(codes, 3, 3), c0, c1),
		},
		[]uint8{
			255 * bool2int(cMask(codes, 0, 0) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 1, 0) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 2, 0) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 3, 0) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 0, 1) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 1, 1) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 2, 1) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 3, 1) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 0, 2) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 1, 2) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 2, 2) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 3, 2) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 0, 3) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 1, 3) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 2, 3) == 3 && c0 <= c1),
			255 * bool2int(cMask(codes, 3, 3) == 3 && c0 <= c1),
		}
}

func bool2int(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func cMask(codes uint32, x, y int) uint8 {
	bt := uint8(8*y + x*2)
	return uint8((codes >> bt) & 0x3)
}

func cSelect(c uint8, c0, c1 uint16) RGB565 {
	switch c {
	case 0:
		return RGB565(c0)
	case 1:
		return RGB565(c1)
	case 2:
		if c0 > c1 {
			r := 2*((c0>>11)&(1<<5-1)) + (c1>>11)&(1<<5-1)
			g := 2*((c0>>5)&(1<<6-1)) + (c1>>5)&(1<<6-1)
			b := 2*(c0&(1<<5-1)) + c1&(1<<5-1)
			return RGB565((r/3)<<11 | (g/3)<<5 | b/3)
		}
		r := (c0>>11)&(1<<5-1) + (c1>>11)&(1<<5-1)
		g := (c0>>5)&(1<<6-1) + (c1>>5)&(1<<6-1)
		b := c0&(1<<5-1) + c1&(1<<5-1)
		return RGB565((r>>1)<<11 | (g>>1)<<5 | b>>1)
	case 3:
		if c0 > c1 {
			r := (c0>>11)&(1<<5-1) + 2*((c1>>11)&(1<<5-1))
			g := (c0>>5)&(1<<6-1) + 2*((c1>>5)&(1<<6-1))
			b := c0&(1<<5-1) + 2*(c1&(1<<5-1))
			return RGB565((r/3)<<11 | (g/3)<<5 | b/3)
		}
		return RGB565(0)
	default:
		panic("invalid code provided (c > 3)")
	}
}

func decodeDtx3Block(b []byte) ([]RGB565, []uint8) {
	if len(b) < 16 {
		panic("not enough data to decode block")
	}

	return decodeDtx1Block(b[8:], true), []uint8{
		(b[7]>>4)<<4 | (b[7] >> 4),
		(b[7]&0x0f)<<4 | (b[7] & 0x0f),
		(b[6]>>4)<<4 | (b[6] >> 4),
		(b[6]&0x0f)<<4 | (b[6] & 0x0f),
		(b[5]>>4)<<4 | (b[5] >> 4),
		(b[5]&0x0f)<<4 | (b[5] & 0x0f),
		(b[4]>>4)<<4 | (b[4] >> 4),
		(b[4]&0x0f)<<4 | (b[4] & 0x0f),
		(b[3]>>4)<<4 | (b[3] >> 4),
		(b[3]&0x0f)<<4 | (b[3] & 0x0f),
		(b[2]>>4)<<4 | (b[2] >> 4),
		(b[2]&0x0f)<<4 | (b[2] & 0x0f),
		(b[1]>>4)<<4 | (b[1] >> 4),
		(b[1]&0x0f)<<4 | (b[1] & 0x0f),
		(b[0]>>4)<<4 | (b[0] >> 4),
		(b[0]&0x0f)<<4 | (b[0] & 0x0f),
	}
}

func decodeDtx5Block(b []byte) ([]RGB565, []uint8) {
	if len(b) < 16 {
		panic("not enough data to decode block")
	}

	a0, a1 := b[0], b[1]
	code := uint64(b[7])<<40 | uint64(b[6])<<32 | uint64(b[5])<<24 | uint64(b[4])<<16 | uint64(b[3])<<8 | uint64(b[2])

	return decodeDtx1Block(b[8:], true), []uint8{
		aBlend(uint8(code>>45)&0x7, a0, a1),
		aBlend(uint8(code>>42)&0x7, a0, a1),
		aBlend(uint8(code>>39)&0x7, a0, a1),
		aBlend(uint8(code>>36)&0x7, a0, a1),
		aBlend(uint8(code>>33)&0x7, a0, a1),
		aBlend(uint8(code>>30)&0x7, a0, a1),
		aBlend(uint8(code>>27)&0x7, a0, a1),
		aBlend(uint8(code>>24)&0x7, a0, a1),
		aBlend(uint8(code>>21)&0x7, a0, a1),
		aBlend(uint8(code>>18)&0x7, a0, a1),
		aBlend(uint8(code>>15)&0x7, a0, a1),
		aBlend(uint8(code>>12)&0x7, a0, a1),
		aBlend(uint8(code>>9)&0x7, a0, a1),
		aBlend(uint8(code>>6)&0x7, a0, a1),
		aBlend(uint8(code>>3)&0x7, a0, a1),
		aBlend(uint8(code>>0)&0x7, a0, a1),
	}
}

func aBlend(code uint8, a0, a1 uint8) uint8 {
	switch code {
	case 0:
		return a0
	case 1:
		return a1
	case 2:
		if a0 > a1 {
			return uint8((6*uint16(a0) + 1*uint16(a1)) / 7)
		}
		return uint8((4*uint16(a0) + 1*uint16(a1)) / 5)
	case 3:
		if a0 > a1 {
			return uint8((5*uint16(a0) + 2*uint16(a1)) / 7)
		}
		return uint8((3*uint16(a0) + 2*uint16(a1)) / 5)
	case 4:
		if a0 > a1 {
			return uint8((4*uint16(a0) + 3*uint16(a1)) / 7)
		}
		return uint8((2*uint16(a0) + 3*uint16(a1)) / 5)
	case 5:
		if a0 > a1 {
			return uint8((3*uint16(a0) + 4*uint16(a1)) / 7)
		}
		return uint8((1*uint16(a0) + 4*uint16(a1)) / 5)
	case 6:
		if a0 > a1 {
			return uint8((2*uint16(a0) + 5*uint16(a1)) / 7)
		}
		return 0
	case 7:
		if a0 > a1 {
			return uint8((1*uint16(a0) + 6*uint16(a1)) / 7)
		}
		return 1
	default:
		panic("invalid code (code > 7)")
	}
}
