package ascii

// implements an image type and color pallete using UTF8 runes
// because.

import (
    "image/color"
    "fmt"
)

// UTF COLORSPACE
// PALETTE DEFINITION
// TODO: better palette, with more UTF codepoints
var txtPallete = []TextColor{ ' ', '.', ':', 'o', 'O', '8', '@', '#', }

// first -- colorspace: TEXT
// mostly a mockery of color.Grey
type TextColor rune
func createPalleteMap (colors []TextColor) map[TextColor]uint32 {
    value := 255 / len(colors)
    mp := make(map[TextColor]uint32, len(colors))

    for i, r := range colors {
        mp[r] = uint32(i * value)
        mp[r] |= mp[r] << 8
    }
    return mp
}
func createPallete (colors []TextColor) color.Palette {
    pal := make([]color.Color, len(colors))
    for i, c := range colors {
        pal[i] = color.Color(c)
    }
    return color.Palette(pal)
}

// precomute uint32 lookups for color conversion
var txtPalleteMap = createPalleteMap(txtPallete)
func (c TextColor) RGBA() (r, g, b, a uint32) {
    y := txtPalleteMap[c]
    return y, y, y, 0xffff
}

// and here's the color space!
var Palette color.Palette = createPallete(txtPallete)


// color model using that palette
func textModel (c color.Color) color.Color {
    if _, ok := c.(TextColor); ok {
        return c
    }
    ret := Palette.Convert(c)
    fmt.Println(ret)
    return ret
}

// The color model for ASCII images
var TextModel color.Model = color.ModelFunc(textModel)
