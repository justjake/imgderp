package derper

import (
	"github.com/justjake/imgderp/ascii"
	"github.com/justjake/imgderp/resize"
	"image"
	"image/gif"
)

type Options struct {
	// affects dimensions
	TargetWidth  int
	TargetHeight int
	PixelRatio   float64

	// affects appearance
	CharSet     string
	CharSetName string
	Invert      bool
}

func Derp(img image.Image, opts *Options) [][]string {
	resizer := getResizer(
		img,
		opts.TargetWidth,
		opts.TargetHeight,
		opts.PixelRatio)
	resized := resizer.ResizeNearestNeighbor()
	colors := getColors(opts)
	txtImage := ascii.Convert(resized, colors)
	frame := txtImage.StringSlice()
	return [][]string{frame}
}

func DerpGif(g *gif.GIF, opts *Options) [][]string {
	colors := getColors(opts)
	resizer := getResizer(
		g.Image[0],
		opts.TargetWidth,
		opts.TargetHeight,
		opts.PixelRatio,
	)
	return encodeFramesSync(g, resizer.TargetWidth, resizer.TargetHeight, colors)
}

var Charsets = map[string][]*ascii.TextColor{
	"default": ascii.DefaultSet,
	"box":     ascii.UnicodeBoxSet,
	"alt":     ascii.AlternateSet,
	"shade":   ascii.UnicodeShadeSet,
	"new":     ascii.SciSet,
}

// creates the resizer for a Derp call
func getResizer(img image.Image, w, h int, pixelRatio float64) *resize.Resizer {
	resizer := resize.NewResizer(img)
	if w == 0 && h == 0 {
		w = 80
	}
	if h == 0 {
		h = resizer.HeightForWidth(w)
	}
	if w == 0 {
		w = resizer.WidthForHeight(h)
	}
	resizer.TargetWidth = w
	resizer.TargetHeight = h
	resizer.TargetHeight = resizer.HeightForPixelRatio(pixelRatio)
	return resizer
}

func getColors(opts *Options) []*ascii.TextColor {
	var colors []*ascii.TextColor

	// preset vs provided
	if opts.CharSet == "" {
		if opts.CharSetName == "" {
			colors = Charsets["default"]
		} else {
			colors = Charsets[opts.CharSetName]
		}
	} else {
		colors = ascii.MakeTextColors([]rune(opts.CharSet)...)
	}

	// flip charset for dark-on-light?
	if opts.Invert {
		colors = ascii.Reverse(colors)
	}

	return colors
}
