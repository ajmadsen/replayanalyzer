package dds

import "io"
import "fmt"

import "bufio"
import "image"
import "image/color"

type reader interface {
	io.Reader
	io.ByteReader
}

type decoder struct {
	r reader

	// header members
	hdrFlags    uint32
	height      uint32
	width       uint32
	pitch       uint32
	depth       uint32
	mipMapCount uint32
	caps        uint32
	caps2       uint32

	// pixel format struct members
	pfFlags     uint32
	fourCC      uint32
	rgbBitCount uint32
	rBitMask    uint32
	gBitMask    uint32
	bBitMask    uint32
	aBitMask    uint32

	rBitShift uint8
	gBitShift uint8
	bBitShift uint8
	aBitShift uint8

	stride int
	line   []byte

	pix       []uint8
	pixStride int
	img       image.Image

	tmp [256]byte
}

// header flags
const (
	// Required in every .dds file.
	DdsdCaps = 0x1
	// Required in every .dds file.
	DdsdHeight = 0x2
	// Required in every .dds file.
	DdsdWidth = 0x4
	// Required when pitch is provided for an uncompressed texture.
	DdsdPitch = 0x8
	// Required in every .dds file.
	DdsdPixelFormat = 0x1000
	// Required in a mipmapped texture.
	DdsdMipMapCount = 0x20000
	// Required when pitch is provided for a compressed texture.
	DdsdLinearSize = 0x80000
	// Required in a depth texture.
	DdsdDepth = 0x800000

	DdsdRequired = DdsdCaps | DdsdHeight | DdsdWidth | DdsdPixelFormat
)

// header caps flags
const (
	// Optional; must be used on any file that contains more than one surface (a mipmap, a cubic environment map, or mipmapped volume texture).
	DdsCapsComplex = 0x8
	// Optional; should be used for a mipmap.
	DdsCapsMipMap = 0x400000
	// Required
	DdsCapsTexture = 0x1000

	DdsSurfaceFlagsMipMap  = DdsCapsComplex | DdsCapsMipMap
	DdsSurfaceFlagsTexture = DdsCapsTexture
	DdsSurfaceFlagsCubeMap = DdsCapsComplex
)

// header caps2 flags
const (
	// Required for a cube map.
	DdsCaps2CubeMap = 0x200
	// Required when these surfaces are stored in a cube map.
	DdsCaps2CubeMapPositiveX = 0x400
	// Required when these surfaces are stored in a cube map.
	DdsCaps2CubeMapNegativeX = 0x800
	// Required when these surfaces are stored in a cube map.
	DdsCaps2CubeMapPositiveY = 0x1000
	// Required when these surfaces are stored in a cube map.
	DdsCaps2CubeMapNegativeY = 0x2000
	// Required when these surfaces are stored in a cube map.
	DdsCaps2CubeMapPositiveZ = 0x4000
	// Required when these surfaces are stored in a cube map.
	DdsCaps2CubeMapNegativeZ = 0x8000
	// Required for a volume texture.
	DdsCaps2Volume = 0x200000

	DdsCubeMapPositiveX = DdsCaps2CubeMap | DdsCaps2CubeMapPositiveX
	DdsCubeMapNegativeX = DdsCaps2CubeMap | DdsCaps2CubeMapNegativeX
	DdsCubeMapPositiveY = DdsCaps2CubeMap | DdsCaps2CubeMapPositiveY
	DdsCubeMapNegativeY = DdsCaps2CubeMap | DdsCaps2CubeMapNegativeY
	DdsCubeMapPositiveZ = DdsCaps2CubeMap | DdsCaps2CubeMapPositiveZ
	DdsCubeMapNegativeZ = DdsCaps2CubeMap | DdsCaps2CubeMapNegativeZ

	DdsCubeMapAllFaces = DdsCubeMapPositiveX | DdsCubeMapNegativeX | DdsCubeMapPositiveY | DdsCubeMapNegativeY | DdsCubeMapPositiveZ | DdsCubeMapNegativeZ

	DdsFlagsVolume = DdsCaps2Volume
)

// pixFmt flags
const (
	// Texture contains alpha data; dwRGBAlphaBitMask contains valid data.
	DdpfAlphaPixels = 0x1
	// Used in some older DDS files for alpha channel only uncompressed data (dwRGBBitCount contains the alpha channel bitcount; dwABitMask contains valid data)
	DdpfAlpha = 0x2
	// Texture contains compressed RGB data; dwFourCC contains valid data.
	DdpfFourCC = 0x4
	// 	Texture contains uncompressed RGB data; dwRGBBitCount and the RGB masks (dwRBitMask, dwGBitMask, dwBBitMask) contain valid data.
	DdpfRgb = 0x40
	// 	Used in some older DDS files for YUV uncompressed data (dwRGBBitCount contains the YUV bit count; dwRBitMask contains the Y mask, dwGBitMask contains the U mask, dwBBitMask contains the V mask)
	DdpfYuv = 0x200
	// Used in some older DDS files for single channel color uncompressed data (dwRGBBitCount contains the luminance channel bit count; dwRBitMask contains the channel mask). Can be combined with DDPF_ALPHAPIXELS for a two channel DDS file.
	DdpfLuminance = 0x20000
)

// known fourCCs
const (
	// DXT1 format
	PixFmtDxt1 = 0x31545844
	// DXT3 format
	PixFmtDxt3 = 0x33545844
	// DXT5 format
	PixFmtDxt5 = 0x35545844
)

const (
	// DdsHeaderSize is the header size in bytes
	ddsHeaderSize = 124
	// PixFmtSize is the pixel format struct size in bytes
	pixFmtSize = 32
)

// Decode decodes a DDS file
func (d *decoder) decode(r io.Reader, configOnly bool) error {
	if rr, ok := r.(reader); ok {
		d.r = rr
	} else {
		d.r = bufio.NewReader(r)
	}

	err := d.readHeader()
	if err != nil {
		return err
	}
	if configOnly {
		return nil
	}

	w := int(d.width+3) / 4
	h := int(d.height+3) / 4

	d.pix = make([]uint8, 4*w*4*h*4)
	d.pixStride = 4 * int(w) * 4

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			err = d.decodeBlock(j*d.pixStride*4 + i*4*4)
			if err != nil {
				return err
			}
		}
	}

	switch d.fourCC {
	case PixFmtDxt1:
		d.img = &image.RGBA{
			Pix:    d.pix,
			Stride: d.pixStride,
			Rect:   image.Rect(0, 0, int(d.width), int(d.height)),
		}
	case PixFmtDxt3, PixFmtDxt5:
		d.img = &image.NRGBA{
			Pix:    d.pix,
			Stride: d.pixStride,
			Rect:   image.Rect(0, 0, int(d.width), int(d.height)),
		}
	}

	return nil
}

func decodeU32LE(dat []byte) uint32 {
	if len(dat) < 4 {
		panic("not enough data to decode uint32")
	}
	return uint32(dat[3])<<24 | uint32(dat[2])<<16 | uint32(dat[1])<<8 | uint32(dat[0])
}

func (d *decoder) readHeader() error {
	_, err := io.ReadFull(d.r, d.tmp[:4])
	if err != nil {
		return err
	}
	if string(d.tmp[:4]) != "DDS " {
		return fmt.Errorf("invalid magic %s", string(d.tmp[:4]))
	}

	_, err = io.ReadFull(d.r, d.tmp[:ddsHeaderSize])
	if err != nil {
		return err
	}

	buf := d.tmp[:124]

	size, buf := decodeU32LE(buf), buf[4:]
	if size != ddsHeaderSize {
		return fmt.Errorf("invalid header size %v", size)
	}

	d.hdrFlags, buf = decodeU32LE(buf), buf[4:]
	if d.hdrFlags&DdsdRequired == 0 {
		return fmt.Errorf("header missing flags  0x%x", (d.hdrFlags^DdsdRequired)&DdsdRequired)
	}

	d.height, buf = decodeU32LE(buf), buf[4:]
	d.width, buf = decodeU32LE(buf), buf[4:]
	d.pitch, buf = decodeU32LE(buf), buf[4:]
	d.depth, buf = decodeU32LE(buf), buf[4:]
	d.mipMapCount, buf = decodeU32LE(buf), buf[4:]

	// skip reserved words
	buf = buf[4*11:]

	// pixel format structure data
	pfSize, buf := decodeU32LE(buf), buf[4:]
	if pfSize != pixFmtSize {
		return fmt.Errorf("invalid pixel format size %v", pfSize)
	}

	d.pfFlags, buf = decodeU32LE(buf), buf[4:]
	d.fourCC, buf = decodeU32LE(buf), buf[4:]
	d.rgbBitCount, buf = decodeU32LE(buf), buf[4:]
	d.rBitMask, buf = decodeU32LE(buf), buf[4:]
	d.gBitMask, buf = decodeU32LE(buf), buf[4:]
	d.bBitMask, buf = decodeU32LE(buf), buf[4:]
	d.aBitMask, buf = decodeU32LE(buf), buf[4:]

	// back to header data
	d.caps, buf = decodeU32LE(buf), buf[4:]
	d.caps2, buf = decodeU32LE(buf), buf[4:]

	return nil
}

func (d *decoder) decodeBlock(offset int) error {
	switch d.fourCC {
	case PixFmtDxt1:
		_, err := io.ReadFull(d.r, d.tmp[:8])
		if err != nil {
			return fmt.Errorf("not enough data to decode block: %v", err)
		}
		decodeDxt1ABlock(d.pix[offset:], d.tmp[:8], d.pixStride)
	case PixFmtDxt3:
		_, err := io.ReadFull(d.r, d.tmp[:16])
		if err != nil {
			return fmt.Errorf("not enough data to decode block: %v", err)
		}
		decodeDxt3Block(d.pix[offset:], d.tmp[:16], d.pixStride)
	case PixFmtDxt5:
		_, err := io.ReadFull(d.r, d.tmp[:16])
		if err != nil {
			return fmt.Errorf("not enough data to decode block: %v", err)
		}
		decodeDxt5Block(d.pix[offset:], d.tmp[:16], d.pixStride)
	default:
		return fmt.Errorf("not a valid fourCC code 0x%x", d.fourCC)
	}
	return nil
}

func (d *decoder) computeBitShifts() {
	if d.pfFlags&DdpfRgb != 0 {
		for ; d.rBitMask&1 != 0; d.rBitMask >>= 1 {
			d.rBitShift++
		}
		for ; d.gBitMask&1 != 0; d.gBitMask >>= 1 {
			d.gBitShift++
		}
		for ; d.bBitMask&1 != 0; d.bBitMask >>= 1 {
			d.bBitShift++
		}
	}
	if d.pfFlags&DdpfAlphaPixels != 0 {
		for ; d.aBitMask&1 != 0; d.aBitMask >>= 1 {
			d.aBitShift++
		}
	}
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	var d decoder
	if err := d.decode(r, true); err != nil {
		return image.Config{}, err
	}
	model := color.NRGBAModel
	if d.fourCC == PixFmtDxt1 {
		model = color.RGBAModel
	}
	return image.Config{
		ColorModel: model,
		Width:      int(d.width),
		Height:     int(d.height),
	}, nil
}

func Decode(r io.Reader) (image.Image, error) {
	var d decoder
	if err := d.decode(r, false); err != nil {
		return &image.RGBA{}, err
	}
	return d.img, nil
}

func init() {
	image.RegisterFormat("dds", "DDS ", Decode, DecodeConfig)
}
