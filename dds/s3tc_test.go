package dds

import (
	"image/color"
	"testing"
)

func TestRGB555Conv(t *testing.T) {
	tests := map[RGB565]color.RGBA{
		RGB565(0xffff): color.RGBA{0xff, 0xff, 0xff, 0xff},
		RGB565(0x3333): color.RGBA{0x31, 0x65, 0x9C, 0xff},
		RGB565(0x8410): color.RGBA{0x84, 0x82, 0x84, 0xff},
		RGB565(0x0000): color.RGBA{0x00, 0x00, 0x00, 0xff},
	}

	for k, v := range tests {
		r1, g1, b1, a1 := k.RGBA()
		r2, g2, b2, a2 := v.RGBA()
		if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
			t.Errorf("converted colors don't match, expected {%b %b %b %b} from %b got {%b %b %b %b}",
				r2, g2, b2, a2, uint16(k), r1, g1, b1, a1)
		}
		if c := RGB565Model.Convert(v); c != k {
			t.Errorf("color model is not accurate expected %v got %v", k, c)
		}
	}
}
