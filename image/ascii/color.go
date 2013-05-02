package ascii

// implements an image type and color pallete using UTF8 runes
// because.

import (
    "image/color"
)

// UTF COLORSPACE
// PALETTE DEFINITION
// TODO: better palette, with more UTF codepoints

// UTF runes as colors
// type TextColor rune
type TextColor struct {
    Rune        rune
    y           uint32
}

func (c *TextColor) RGBA() (r, g, b, a uint32) {
    y := c.y
    return y, y, y, 0xffff
}

func (c *TextColor) String() string {
    return string(c.Rune)
}

func Reverse(cs []*TextColor) []*TextColor {
    runes := make([]rune, len(cs))
    for i, n := 0, len(cs); i < n; i++ {
        runes[i] = cs[n-1-i].Rune
    }
    return MakeTextColors(runes...)
}

// Make a new palette. Only way to get text colors.
// color values per character are global, so overlapping palettes will break things
// TODO: rework TextColor as a struct to do local lookups
func MakeTextColors(chars ...rune) []*TextColor {
    value := 255 / len(chars)

    pal := make([]*TextColor, len(chars))

    for i, r := range chars {
        y := uint32(i * value)
        y |= y << 8
        c := TextColor{r, y}
        pal[i] = &c
    }

    return pal
}

// slice of *TextColor -> color.Palette
func NewPalette(p []*TextColor) color.Palette {
    pal := make([]color.Color, len(p))
    for i, r := range p {
        pal[i] = color.Color(r)
    }
    return pal
}

// default Palette
var (
    DefaultSet      = MakeTextColors([]rune(" .:oO8@")...)
    AlternateSet    = MakeTextColors([]rune(" .:;+=xX$&")...)
    SciSet          = MakeTextColors([]rune(" .,-+=:;/XHM%@#$")...)
    UnicodeBoxSet   = MakeTextColors([]rune(" ▏▎▍▌▋▊")...)
    UnicodeShadeSet = MakeTextColors([]rune(" .░▒▓￭")...)
)

// The color model for ASCII images
var TextModel color.Model = color.Model(NewPalette(DefaultSet))
