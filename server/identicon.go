package server

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strconv"
	"time"
	"unicode/utf8"
)

type Identicon struct {
	hash string
	img  *image.RGBA
}

type IdenticonOptions struct {
	background color.RGBA
	hash       string
	margin     float64
	size       int
}

func defaultIdenticonOptions() *IdenticonOptions {
	return &IdenticonOptions{color.RGBA{240, 240, 240, 255}, "", 0.08, 64}
}

func NewIdenticon(hash string, options *IdenticonOptions) *Identicon {
	if options == nil {
		options = defaultIdenticonOptions()
	}
	hueInt, _ := strconv.ParseInt(hash[len(hash)-8:], 16, 64)
	hue := float64(hueInt % 360)
	foreground := hsl2rgb(hue, 0.5, 0.7)
	img := image.NewRGBA(image.Rect(0, 0, options.size, options.size))
	draw.Draw(img, img.Bounds(), &image.Uniform{options.background}, image.ZP, draw.Src)
	baseMargin := int(float64(options.size) * options.margin)
	cell := int((options.size - baseMargin*2) / 5)
	margin := int((options.size - cell*5) / 2)
	var color color.RGBA
	pt := func(k int) int { return k*cell + margin }
	for i := 0; i < 15; i++ {
		if val, _ := strconv.ParseInt(string(hash[i]), 16, 16); val%2 == 0 {
			color = *foreground
		} else {
			color = options.background
		}
		if i < 5 {
			draw.Draw(img, image.Rect(pt(2), pt(i), pt(2)+cell, pt(i)+cell), &image.Uniform{color},
				image.ZP, draw.Src)
		} else if i < 10 {
			draw.Draw(img, image.Rect(pt(1), pt(i-5), pt(1)+cell, pt(i-5)+cell), &image.Uniform{color},
				image.ZP, draw.Src)
			draw.Draw(img, image.Rect(pt(3), pt(i-5), pt(3)+cell, pt(i-5)+cell), &image.Uniform{color},
				image.ZP, draw.Src)
		} else if i < 15 {
			draw.Draw(img, image.Rect(pt(0), pt(i-10), pt(0)+cell, pt(i-10)+cell), &image.Uniform{color},
				image.ZP, draw.Src)
			draw.Draw(img, image.Rect(pt(4), pt(i-10), pt(4)+cell, pt(i-10)+cell), &image.Uniform{color},
				image.ZP, draw.Src)
		}
	}
	return &Identicon{hash, img}
}

func (I *Identicon) Save(filepath string) error {
	fp, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()
	png.Encode(fp, I.img)
	return nil
}

func (I *Identicon) ToBase64() string {
	return ""
}

func GenerateHash(input string) string {
	hash := 0
	salt := time.Now().String()
	input = salt + input
	n := utf8.RuneCountInString(input)
	for i := 0; i < n; i++ {
		chr := rune(input[i])
		hash = ((hash << 5) - hash) + int(chr)
		hash = hash | 0
	}
	return strconv.FormatInt(int64(hash), 16)
}

func LoadIdenticon(filepath string) (string, error) {
	return "", nil
}

func hsl2rgb(h, s, l float64) *color.RGBA {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(float64(int(h/60)%2-1)))
	m := l - c/2
	var rp, gp, bp float64
	if h < 60 {
		rp, gp, bp = c, x, 0
	} else if h < 120 {
		rp, gp, bp = x, c, 0
	} else if h < 180 {
		rp, gp, bp = 0, c, x
	} else if h < 240 {
		rp, gp, bp = 0, x, c
	} else if h < 300 {
		rp, gp, bp = x, 0, c
	} else {
		rp, gp, bp = c, 0, x
	}
	r := touint8((rp + m) * 255)
	g := touint8((gp + m) * 255)
	b := touint8((bp + m) * 255)
	return &color.RGBA{r, g, b, 255}
}

func touint8(x float64) uint8 {
	if x > 0 {
		return uint8(x + 0.5)
	}
	return uint8(x - 0.5)
}
