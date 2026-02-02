package base64Captcha

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"math/big"
)

const (
	digitFontWidth     = 11
	digitFontHeight    = 18
	digitFontBlackChar = 1
)

// ItemDigit digits captcha Struct
type ItemDigit struct {
	width  int
	height int
	*image.Paletted
	dotSize  int
	dotCount int
	maxSkew  float64
	//rng      siprng
}

// NewItemDigit create a instance of item-digit
func NewItemDigit(width int, height int, dotCount int, maxSkew float64) (*ItemDigit, error) {
	itemDigit := &ItemDigit{width: width, height: height, dotCount: dotCount, maxSkew: maxSkew}
	//init image.Paletted
	colorPalette, err := createRandPaletteColors(dotCount)
	if err != nil {
		return nil, err
	}
	itemDigit.Paletted = image.NewPaletted(image.Rect(0, 0, width, height), colorPalette)
	return itemDigit, nil
}

func createRandPaletteColors(dotCount int) (color.Palette, error) {
	p := make([]color.Color, dotCount+1)
	// Transparent color.
	p[0] = color.RGBA{0xFF, 0xFF, 0xFF, 0x00}
	// Primary color.
	GN, err := rand.Int(rand.Reader, big.NewInt(129))
	if err != nil {
		return nil, err
	}
	green := int(GN.Int64())
	RN, err := rand.Int(rand.Reader, big.NewInt(129))
	if err != nil {
		return nil, err
	}
	red := int(RN.Int64())
	BN, err := rand.Int(rand.Reader, big.NewInt(129))
	if err != nil {
		return nil, err
	}
	blue := int(BN.Int64())
	prim := color.RGBA{
		uint8(red),
		uint8(green),
		uint8(blue),
		0xFF,
	}

	if dotCount == 0 {
		p[0] = prim
		return p, fmt.Errorf("dotCount must be greater than 0")
	}

	p[1] = prim
	// Circle colors.
	for i := 2; i <= dotCount; i++ {
		p[i], err = randomBrightness(prim, 255)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (m *ItemDigit) calculateSizes(width, height, ncount int) {
	// Goal: fit all digits inside the image.
	var border int
	if width > height {
		border = height / 4
	} else {
		border = width / 4
	}
	// Convert everything to floats for calculations.
	w := float64(width - border*2)
	h := float64(height - border*2)
	// fw takes into account 1-dot spacing between digits.
	fw := float64(digitFontWidth + 1)
	fh := float64(digitFontHeight)
	nc := float64(ncount)
	// Calculate the width of a single digit taking into account only the
	// width of the image.
	nw := w / nc
	// Calculate the height of a digit from this width.
	nh := nw * fh / fw
	// Digit too high?
	if nh > h {
		// Fit digits based on height.
		nh = h
		nw = fw / fh * nh
	}
	// Calculate dot size.
	m.dotSize = int(nh / fh)
	if m.dotSize < 1 {
		m.dotSize = 1
	}
	// Save everything, making the actual width smaller by 1 dot to account
	// for spacing between digits.
	m.width = int(nw) - m.dotSize
	m.height = int(nh)
}

func (m *ItemDigit) drawHorizLine(fromX, toX, y int, colorIdx uint8) {
	for x := fromX; x <= toX; x++ {
		m.SetColorIndex(x, y, colorIdx)
	}
}

func (m *ItemDigit) drawCircle(x, y, radius int, colorIdx uint8) {
	f := 1 - radius
	dfx := 1
	dfy := -2 * radius
	xo := 0
	yo := radius

	m.SetColorIndex(x, y+radius, colorIdx)
	m.SetColorIndex(x, y-radius, colorIdx)
	m.drawHorizLine(x-radius, x+radius, y, colorIdx)

	for xo < yo {
		if f >= 0 {
			yo--
			dfy += 2
			f += dfy
		}
		xo++
		dfx += 2
		f += dfx
		m.drawHorizLine(x-xo, x+xo, y+yo, colorIdx)
		m.drawHorizLine(x-xo, x+xo, y-yo, colorIdx)
		m.drawHorizLine(x-yo, x+yo, y+xo, colorIdx)
		m.drawHorizLine(x-yo, x+yo, y-xo, colorIdx)
	}
}

func (m *ItemDigit) fillWithCircles(n, maxradius int) error {
	maxx := m.Bounds().Max.X
	maxy := m.Bounds().Max.Y
	for i := 0; i < n; i++ {
		//colorIdx := uint8(m.rng.Int(1, m.dotCount-1))
		colorIdx, err := randIntRange(1, m.dotCount-1)
		if err != nil {
			return err
		}
		//r := m.rng.Int(1, maxradius)
		r, err := randIntRange(1, maxradius)
		if err != nil {
			return err
		}
		//m.drawCircle(m.rng.Int(r, maxx-r), m.rng.Int(r, maxy-r), r, colorIdx)
		x, err := randIntRange(r, maxx-r)
		if err != nil {
			return err
		}
		y, err := randIntRange(r, maxy-r)
		if err != nil {
			return err
		}
		m.drawCircle(x, y, r, uint8(colorIdx))
	}
	return nil
}

func (m *ItemDigit) strikeThrough() error {
	maxx := m.Bounds().Max.X
	maxy := m.Bounds().Max.Y
	y, err := randIntRange(maxy/3, maxy-maxy/3)
	if err != nil {
		return err
	}
	amplitude, err := randFloat64Range(5, 20)
	if err != nil {
		return err
	}
	period, err := randFloat64Range(80, 180)
	if err != nil {
		return err
	}
	dx := 2.0 * math.Pi / period
	for x := 0; x < maxx; x++ {
		xo := amplitude * math.Cos(float64(y)*dx)
		yo := amplitude * math.Sin(float64(x)*dx)
		for yn := 0; yn < m.dotSize; yn++ {
			rN, err := rand.Int(rand.Reader, big.NewInt(int64(m.dotSize)))
			if err != nil {
				return err
			}
			r := int(rN.Int64())
			m.drawCircle(x+int(xo), y+int(yo)+(yn*m.dotSize), r/2, 1)
		}
	}
	return nil
}

// draw digit
func (m *ItemDigit) drawDigit(digit []byte, x, y int) error {
	skf, err := randFloat64Range(-m.maxSkew, m.maxSkew)
	if err != nil {
		return err
	}
	xs := float64(x)
	r := m.dotSize / 2
	ySum, err := randIntRange(-r, r)
	if err != nil {
		return err
	}
	y += ySum
	for yo := 0; yo < digitFontHeight; yo++ {
		for xo := 0; xo < digitFontWidth; xo++ {
			if digit[yo*digitFontWidth+xo] != digitFontBlackChar {
				continue
			}
			m.drawCircle(x+xo*m.dotSize, y+yo*m.dotSize, r, 1)
		}
		xs += skf
		x = int(xs)
	}
	return nil
}

func (m *ItemDigit) distort(amplude float64, period float64) {
	w := m.Bounds().Max.X
	h := m.Bounds().Max.Y

	oldm := m.Paletted
	newm := image.NewPaletted(image.Rect(0, 0, w, h), oldm.Palette)

	dx := 2.0 * math.Pi / period
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			xo := amplude * math.Sin(float64(y)*dx)
			yo := amplude * math.Cos(float64(x)*dx)
			newm.SetColorIndex(x, y, oldm.ColorIndexAt(x+int(xo), y+int(yo)))
		}
	}
	m.Paletted = newm
}

func randomBrightness(c color.RGBA, max uint8) (color.RGBA, error) {
	minc := min3(c.R, c.G, c.B)
	maxc := max3(c.R, c.G, c.B)
	if maxc > max {
		return c, nil
	}
	// n := rand.Intn(int(max-maxc)) - int(minc)
	nN, err := rand.Int(rand.Reader, big.NewInt(int64(int(max-maxc)+1+int(minc))))
	if err != nil {
		return color.RGBA{}, err
	}
	n := int(nN.Int64()) - int(minc)
	return color.RGBA{
		uint8(int(c.R) + n),
		uint8(int(c.G) + n),
		uint8(int(c.B) + n),
		uint8(c.A),
	}, nil
}

func min3(x, y, z uint8) (m uint8) {
	m = x
	if y < m {
		m = y
	}
	if z < m {
		m = z
	}
	return
}

func max3(x, y, z uint8) (m uint8) {
	m = x
	if y > m {
		m = y
	}
	if z > m {
		m = z
	}
	return
}

// EncodeBinary encodes an image to PNG and returns a byte slice.
func (m *ItemDigit) EncodeBinary() []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, m.Paletted); err != nil {
		panic(err.Error())
	}
	return buf.Bytes()
}

// WriteTo writes captcha character in png format into the given io.Writer, and
// returns the number of bytes written and an error if any.
func (m *ItemDigit) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(m.EncodeBinary())
	return int64(n), err
}

// EncodeB64string encodes an image to base64 string
func (m *ItemDigit) EncodeB64string() string {
	return fmt.Sprintf("data:%s;base64,%s", MimeTypeImage, base64.StdEncoding.EncodeToString(m.EncodeBinary()))
}
