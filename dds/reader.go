package dds

import "io"
import "fmt"

import "bufio"
import "image"

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
	// only handle this format for now
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

	_, err = io.ReadFull(d.r, d.tmp[:4])
	if err != nil {
		return err
	}

	return nil
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	var d decoder
	if err := d.decode(r, true); err != nil {
		return image.Config{}, err
	}
	return image.Config{
		ColorModel: RGB565Model,
		Width:      int(d.width),
		Height:     int(d.height),
	}, nil
}