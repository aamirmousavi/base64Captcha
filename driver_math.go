package base64Captcha

import (
	"crypto/rand"
	"fmt"
	"image/color"
	"math/big"
	"strings"

	"github.com/golang/freetype/truetype"
)

// DriverMath captcha config for captcha math
type DriverMath struct {
	//Height png height in pixel.
	Height int

	// Width Captcha png width in pixel.
	Width int

	//NoiseCount text noise count.
	NoiseCount int

	//ShowLineOptions := OptionShowHollowLine | OptionShowSlimeLine | OptionShowSineLine .
	ShowLineOptions int

	//BgColor captcha image background color (optional)
	BgColor *color.RGBA

	//fontsStorage font storage (optional)
	fontsStorage FontsStorage

	//Fonts loads by name see fonts.go's comment
	Fonts      []string
	fontsArray []*truetype.Font
}

// NewDriverMath creates a driver of math
func NewDriverMath(height int, width int, noiseCount int, showLineOptions int, bgColor *color.RGBA, fontsStorage FontsStorage, fonts []string) *DriverMath {
	if fontsStorage == nil {
		fontsStorage = DefaultEmbeddedFonts
	}

	tfs := []*truetype.Font{}
	for _, fff := range fonts {
		tf := fontsStorage.LoadFontByName("fonts/" + fff)
		tfs = append(tfs, tf)
	}

	if len(tfs) == 0 {
		tfs = fontsAll
	}

	return &DriverMath{Height: height, Width: width, NoiseCount: noiseCount, ShowLineOptions: showLineOptions, fontsArray: tfs, BgColor: bgColor, Fonts: fonts}
}

// ConvertFonts loads fonts from names
func (d *DriverMath) ConvertFonts() *DriverMath {
	if d.fontsStorage == nil {
		d.fontsStorage = DefaultEmbeddedFonts
	}

	tfs := []*truetype.Font{}
	for _, fff := range d.Fonts {
		tf := d.fontsStorage.LoadFontByName("fonts/" + fff)
		tfs = append(tfs, tf)
	}
	if len(tfs) == 0 {
		tfs = fontsAll
	}
	d.fontsArray = tfs

	return d
}

// GenerateIdQuestionAnswer creates id,captcha content and answer
func (d *DriverMath) GenerateIdQuestionAnswer() (id, question, answer string, _ error) {
	id = RandomId()
	operators := []string{"+", "-", "x"}
	var mathResult int32
	n, err := rand.Int(rand.Reader, big.NewInt(3))
	if err != nil {
		return "", "", "", err
	}
	switch operators[int(n.Int64())] {
	case "+":
		aN, err := rand.Int(rand.Reader, big.NewInt(20))
		if err != nil {
			return "", "", "", err
		}
		a := int32(aN.Int64())
		bN, err := rand.Int(rand.Reader, big.NewInt(20))
		if err != nil {
			return "", "", "", err
		}
		b := int32(bN.Int64())
		question = fmt.Sprintf("%d+%d=?", a, b)
		mathResult = a + b
	case "x":
		aN, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", "", "", err
		}
		a := int32(aN.Int64())
		bN, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", "", "", err
		}
		b := int32(bN.Int64())
		question = fmt.Sprintf("%dx%d=?", a, b)
		mathResult = a * b
	default:
		aNFirst, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			return "", "", "", err
		}
		aFirst := int32(aNFirst.Int64())
		aNSecond, err := rand.Int(rand.Reader, big.NewInt(20))
		if err != nil {
			return "", "", "", err
		}
		aSecond := int32(aNSecond.Int64())
		a := aFirst + aSecond

		bN, err := rand.Int(rand.Reader, big.NewInt(int64(a)))
		if err != nil {
			return "", "", "", err
		}
		b := int32(bN.Int64())
		question = fmt.Sprintf("%d-%d=?", a, b)
		mathResult = a - b

	}
	answer = fmt.Sprintf("%d", mathResult)
	return
}

// DrawCaptcha creates math captcha item
func (d *DriverMath) DrawCaptcha(question string) (item Item, _ error) {
	var bgc color.RGBA
	if d.BgColor != nil {
		bgc = *d.BgColor
	} else {
		var err error
		bgc, err = RandLightColor()
		if err != nil {
			return nil, err
		}
	}
	itemChar := NewItemChar(d.Width, d.Height, bgc)

	//波浪线 比较丑
	if d.ShowLineOptions&OptionShowHollowLine == OptionShowHollowLine {
		itemChar.drawHollowLine()
	}

	//背景有文字干扰
	if d.NoiseCount > 0 {
		noise, err := RandText(d.NoiseCount, strings.Repeat(TxtNumbers, d.NoiseCount))
		if err != nil {
			return nil, err
		}
		err = itemChar.drawNoise(noise, fontsAll)
		if err != nil {
			return nil, err
		}
	}

	//画 细直线 (n 条)
	if d.ShowLineOptions&OptionShowSlimeLine == OptionShowSlimeLine {
		itemChar.drawSlimLine(3)
	}

	//画 多个小波浪线
	if d.ShowLineOptions&OptionShowSineLine == OptionShowSineLine {
		itemChar.drawSineLine()
	}

	//draw question
	err := itemChar.drawText(question, d.fontsArray)
	if err != nil {
		return nil, err
	}
	return itemChar, nil
}
