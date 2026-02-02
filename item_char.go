package base64Captcha

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"math/big"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

// ItemChar captcha item of unicode characters
type ItemChar struct {
	bgColor color.Color
	width   int
	height  int
	nrgba   *image.NRGBA
}

// NewItemChar creates a captcha item of characters
func NewItemChar(w int, h int, bgColor color.RGBA) *ItemChar {
	d := ItemChar{width: w, height: h}
	m := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(m, m.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)
	d.nrgba = m
	return &d
}

// drawHollowLine draw strong and bold white line.
func (item *ItemChar) drawHollowLine() (*ItemChar, error) {

	first := item.width / 20
	end := first * 19

	lineColor, err := RandLightColor()
	if err != nil {
		return item, err
	}

	// x1 := float64(rand.Intn(first))
	x1n, err := rand.Int(rand.Reader, big.NewInt(int64(first)))
	if err != nil {
		return item, err
	}
	x1 := float64(x1n.Int64())
	//y1 := float64(rand.Intn(y)+y);

	// x2 := float64(rand.Intn(first) + end)
	x2n, err := rand.Int(rand.Reader, big.NewInt(int64(first)))
	if err != nil {
		return item, err
	}
	x2 := float64(x2n.Int64() + int64(end))
	//y2 := float64(rand.Intn(y)+y);

	// multiple := float64(rand.Intn(5)+3) / float64(5)
	multipleN, err := rand.Int(rand.Reader, big.NewInt(8))
	if err != nil {
		return item, err
	}
	multiple := float64(multipleN.Int64()+3) / float64(5)
	if int(multiple*10)%3 == 0 {
		multiple = multiple * -1.0
	}

	w := item.height / 20

	for ; x1 < x2; x1++ {

		y := math.Sin(x1*math.Pi*multiple/float64(item.width)) * float64(item.height/3)

		if multiple < 0 {
			y = y + float64(item.height/2)
		}
		item.nrgba.Set(int(x1), int(y), lineColor)

		for i := 0; i <= w; i++ {
			item.nrgba.Set(int(x1), int(y)+i, lineColor)
		}
	}

	return item, nil
}

// drawSineLine draw a sine line.
func (item *ItemChar) drawSineLine() (*ItemChar, error) {
	var py float64

	//振幅
	aN, err := rand.Int(rand.Reader, big.NewInt(int64(item.height/2)))
	if err != nil {
		return item, err
	}
	a := int(aN.Int64())

	//Y轴方向偏移量
	b, err := random(int64(-item.height/4), int64(item.height/4))
	if err != nil {
		return item, err
	}

	//X轴方向偏移量
	f, err := random(int64(-item.height/4), int64(item.height/4))
	if err != nil {
		return item, err
	}
	// 周期
	var t float64
	if item.height > item.width/2 {
		t, err = random(int64(item.width/2), int64(item.height))
		if err != nil {
			return item, err
		}
	} else if item.height == item.width/2 {
		t = float64(item.height)
	} else {
		t, err = random(int64(item.height), int64(item.width/2))
		if err != nil {
			return item, err
		}
	}
	w := float64((2 * math.Pi) / t)

	// 曲线横坐标起始位置
	px1 := 0
	px2, err := random(int64(float64(item.width)*0.8), int64(item.width))

	c, err := RandDeepColor()
	if err != nil {
		return item, err
	}

	for px := px1; px < int(px2); px++ {
		if w != 0 {
			py = float64(a)*math.Sin(w*float64(px)+f) + b + (float64(item.width) / float64(5))
			i := item.height / 5
			for i > 0 {
				item.nrgba.Set(px+i, int(py), c)
				//fmt.Println(px + i,int(py) )
				i--
			}
		}
	}

	return item, nil
}

// drawSlimLine draw n slim-random-color lines.
func (item *ItemChar) drawSlimLine(num int) (*ItemChar, error) {

	first := item.width / 10
	end := first * 9

	y := item.height / 3

	for i := 0; i < num; i++ {

		// point1 := point{X: rand.Intn(first), Y: rand.Intn(y)}
		x1n, err := rand.Int(rand.Reader, big.NewInt(int64(first)))
		if err != nil {
			return item, err
		}
		y1n, err := rand.Int(rand.Reader, big.NewInt(int64(y)))
		if err != nil {
			return item, err
		}
		point1 := point{X: int(x1n.Int64()), Y: int(y1n.Int64())}

		// point2 := point{X: rand.Intn(first) + end, Y: rand.Intn(y)}
		x2n, err := rand.Int(rand.Reader, big.NewInt(int64(first)))
		if err != nil {
			return item, err
		}
		y2n, err := rand.Int(rand.Reader, big.NewInt(int64(y)))
		if err != nil {
			return item, err
		}
		point2 := point{X: int(x2n.Int64()) + end, Y: int(y2n.Int64())}

		if i%2 == 0 {
			// point1.Y = rand.Intn(y) + y*2
			y1n, err := rand.Int(rand.Reader, big.NewInt(int64(y)))
			if err != nil {
				return item, err
			}
			point1.Y = int(y1n.Int64()) + y*2
			// point2.Y = rand.Intn(y)
			y2n, err := rand.Int(rand.Reader, big.NewInt(int64(y)))
			if err != nil {
				return item, err
			}
			point2.Y = int(y2n.Int64())
		} else {
			// point1.Y = rand.Intn(y) + y*(i%2)
			y1n, err := rand.Int(rand.Reader, big.NewInt(int64(y)))
			if err != nil {
				return item, err
			}
			point1.Y = int(y1n.Int64()) + y*(i%2)
			// point2.Y = rand.Intn(y) + y*2
			y2n, err := rand.Int(rand.Reader, big.NewInt(int64(y)))
			if err != nil {
				return item, err
			}
			point2.Y = int(y2n.Int64()) + y*2
		}

		randDeepColor, err := RandDeepColor()
		if err != nil {
			return item, err
		}
		item.drawBeeline(point1, point2, randDeepColor)

	}
	return item, nil
}

func (item *ItemChar) drawBeeline(point1 point, point2 point, lineColor color.RGBA) {
	dx := math.Abs(float64(point1.X - point2.X))

	dy := math.Abs(float64(point2.Y - point1.Y))
	sx, sy := 1, 1
	if point1.X >= point2.X {
		sx = -1
	}
	if point1.Y >= point2.Y {
		sy = -1
	}
	err := dx - dy
	for {
		item.nrgba.Set(point1.X, point1.Y, lineColor)
		item.nrgba.Set(point1.X+1, point1.Y, lineColor)
		item.nrgba.Set(point1.X-1, point1.Y, lineColor)
		item.nrgba.Set(point1.X+2, point1.Y, lineColor)
		item.nrgba.Set(point1.X-2, point1.Y, lineColor)
		if point1.X == point2.X && point1.Y == point2.Y {
			return
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			point1.X += sx
		}
		if e2 < dx {
			err += dx
			point1.Y += sy
		}
	}
}

func (item *ItemChar) drawNoise(noiseText string, fonts []*truetype.Font) error {

	c := freetype.NewContext()
	c.SetDPI(imageStringDpi)

	c.SetClip(item.nrgba.Bounds())
	c.SetDst(item.nrgba)
	c.SetHinting(font.HintingFull)
	// rawFontSize := float64(item.height) / (1 + float64(rand.Intn(7))/float64(10))
	rfsN, err := rand.Int(rand.Reader, big.NewInt(7))
	if err != nil {
		return err
	}
	rawFontSize := float64(item.height) / (1 + float64(rfsN.Int64())/float64(10))

	for _, char := range noiseText {
		rwN, err := rand.Int(rand.Reader, big.NewInt(int64(item.width)))
		if err != nil {
			return err
		}
		rw := int(rwN.Int64())
		rhN, err := rand.Int(rand.Reader, big.NewInt(int64(item.height)))
		if err != nil {
			return err
		}
		rh := int(rhN.Int64())
		fsN, err := rand.Int(rand.Reader, big.NewInt(5))
		if err != nil {
			return err
		}
		fontSize := rawFontSize/2 + float64(fsN.Int64())
		randLightColor, err := RandLightColor()
		if err != nil {
			return err
		}
		c.SetSrc(image.NewUniform(randLightColor))
		c.SetFontSize(fontSize)
		randFonts, err := randFontFrom(fonts)
		if err != nil {
			return err
		}
		c.SetFont(randFonts)
		pt := freetype.Pt(rw, rh)
		if _, err := c.DrawString(string(char), pt); err != nil {
			log.Println(err)
		}
	}
	return nil
}

//drawText draw captcha string to image.把文字写入图像验证码

func (item *ItemChar) drawText(text string, fonts []*truetype.Font) error {
	c := freetype.NewContext()
	c.SetDPI(imageStringDpi)
	c.SetClip(item.nrgba.Bounds())
	c.SetDst(item.nrgba)
	c.SetHinting(font.HintingFull)

	if len(text) == 0 {
		return errors.New("text must not be empty, there is nothing to draw")
	}

	fontWidth := item.width / len(text)

	for i, s := range text {
		fsN, err := rand.Int(rand.Reader, big.NewInt(7))
		if err != nil {
			return err
		}
		fontSize := item.height * (int(fsN.Int64()) + 7) / 16
		src, err := RandDeepColor()
		if err != nil {
			return err
		}
		c.SetSrc(image.NewUniform(src))
		c.SetFontSize(float64(fontSize))
		randFont, err := randFontFrom(fonts)
		if err != nil {
			return err
		}
		c.SetFont(randFont)
		x := fontWidth*i + fontWidth/fontSize
		rhN, err := rand.Int(rand.Reader, big.NewInt(int64(item.height/16*3)))
		if err != nil {
			return err
		}
		y := item.height/2 + fontSize/2 - int(rhN.Int64())
		pt := freetype.Pt(x, y)
		if _, err := c.DrawString(string(s), pt); err != nil {
			return err
		}
	}
	return nil
}

// BinaryEncoding encodes an image to PNG and returns a byte slice.
func (item *ItemChar) BinaryEncoding() []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, item.nrgba); err != nil {
		panic(err.Error())
	}
	return buf.Bytes()
}

// WriteTo writes captcha character in png format into the given io.Writer, and
// returns the number of bytes written and an error if any.
func (item *ItemChar) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(item.BinaryEncoding())
	return int64(n), err
}

// EncodeB64string encodes an image to base64 string
func (item *ItemChar) EncodeB64string() string {
	return fmt.Sprintf("data:%s;base64,%s", MimeTypeImage, base64.StdEncoding.EncodeToString(item.BinaryEncoding()))
}

type point struct {
	X int
	Y int
}
