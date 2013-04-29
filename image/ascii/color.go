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
    Rune rune
    lookupTable *map[rune]uint32
}

func (c *TextColor) RGBA() (r, g, b, a uint32) {
    y := (*c.lookupTable)[ c.Rune ]
    return y, y, y, 0xffff
}

func (c *TextColor) String() string {
    return string(c.Rune)
}


// Make a new palette. Only way to get text colors.
// color values per character are global, so overlapping palettes will break things
// TODO: rework TextColor as a struct to do local lookups
func MakePalette (chars... rune) color.Palette {
    value := 255 / len(chars)

    txtPalleteMap := make(map[rune]uint32, len(chars))
    pal := make([]color.Color, len(chars))

    for i, r := range chars {
        txtPalleteMap[r] = uint32(i * value)
        txtPalleteMap[r] |= txtPalleteMap[r] << 8

        c := TextColor{r, &txtPalleteMap}
        pal[i] = &c
    }

    return pal
}

// default Palette
var Palette color.Palette = MakePalette(' ', '.', ':', 'o', 'O', '8', '@')

// The color model for ASCII images
var TextModel color.Model = color.Model(Palette)
