package base64Captcha

import (
	"crypto/rand"
	"image/color"
	"math/big"
	"strings"

	"github.com/golang/freetype/truetype"
)

// DriverChinese is a driver of unicode Chinese characters.
type DriverChinese struct {
	//Height png height in pixel.
	Height int
	//Width Captcha png width in pixel.
	Width int

	//NoiseCount text noise count.
	NoiseCount int

	//ShowLineOptions := OptionShowHollowLine | OptionShowSlimeLine | OptionShowSineLine .
	ShowLineOptions int

	//Length random string length.
	Length int

	//Source is a unicode which is the rand string from.
	Source string

	//BgColor captcha image background color (optional)
	BgColor *color.RGBA

	//fontsStorage font storage (optional)
	fontsStorage FontsStorage

	//Fonts loads by name see fonts.go's comment
	Fonts      []string
	fontsArray []*truetype.Font
}

// NewDriverChinese creates a driver of Chinese characters
func NewDriverChinese(height int, width int, noiseCount int, showLineOptions int, length int, source string, bgColor *color.RGBA, fontsStorage FontsStorage, fonts []string) *DriverChinese {
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

	return &DriverChinese{Height: height, Width: width, NoiseCount: noiseCount, ShowLineOptions: showLineOptions, Length: length, Source: source, BgColor: bgColor, fontsStorage: fontsStorage, fontsArray: tfs}
}

// ConvertFonts loads fonts by names
func (d *DriverChinese) ConvertFonts() *DriverChinese {
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

// GenerateIdQuestionAnswer generates captcha content and its answer
func (d *DriverChinese) GenerateIdQuestionAnswer() (id, content, answer string, _ error) {
	id = RandomId()

	ss := strings.Split(d.Source, ",")
	length := len(ss)
	if length == 1 {
		c, err := RandText(d.Length, ss[0])
		if err != nil {
			return "", "", "", err
		}
		return id, c, c, nil
	}
	if length <= d.Length {
		c, err := RandText(d.Length, TxtNumbers+TxtAlphabet)
		if err != nil {
			return "", "", "", err
		}
		return id, c, c, nil
	}

	res := make([]string, d.Length)
	for k := range res {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
		if err != nil {
			return "", "", "", err
		}
		res[k] = ss[int(n.Int64())]
	}

	content = strings.Join(res, "")
	return id, content, content, nil
}

// DrawCaptcha generates captcha item(image)
func (d *DriverChinese) DrawCaptcha(content string) (item Item, _ error) {

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

	//draw hollow line
	if d.ShowLineOptions&OptionShowHollowLine == OptionShowHollowLine {
		itemChar.drawHollowLine()
	}

	//draw slime line
	if d.ShowLineOptions&OptionShowSlimeLine == OptionShowSlimeLine {
		itemChar.drawSlimLine(3)
	}

	//draw sine line
	if d.ShowLineOptions&OptionShowSineLine == OptionShowSineLine {
		itemChar.drawSineLine()
	}

	//draw noise
	if d.NoiseCount > 0 {
		source := TxtNumbers + TxtAlphabet + ",.[]<>"
		noise, err := RandText(d.NoiseCount, strings.Repeat(source, d.NoiseCount))
		if err != nil {
			return
		}
		err = itemChar.drawNoise(noise, d.fontsArray)
		if err != nil {
			return
		}
	}

	//draw content
	err := itemChar.drawText(content, d.fontsArray)
	if err != nil {
		return nil, err
	}

	return itemChar, nil
}
