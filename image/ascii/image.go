package ascii

// implements an image type and color pallete using UTF8 runes
// because.

import (
    "image"
    "image/color"
    "io" // for a silly reason ;)
    "strings"
)

// TextImage type!

// cols[rows[letters]]
type Image [][]*TextColor

func (img *Image) ColorModel() color.Model {
    return TextModel
}

func (img *Image) Bounds() image.Rectangle {
    h := len(*img)
    w := len((*img)[0])

    return image.Rect(0, 0, w, h)
}

func (img *Image) At(x, y int) color.Color {
    return ((*img)[y][x])
}

func (t *Image) String() string {
    grid := *t
    lines := make([]string, len(grid))
    for i, v := range grid {
        rs := make([]rune, len(v))
        for k := range rs {
            rs[k] = v[k].Rune
        }
        lines[i] = string(rs) + "\n"
    }
    return strings.Join(lines, "")
}

// Image creation
func NewImage(w, h uint) *Image {
    size := w * h
    store := make([]*TextColor, size)
    grid := make([][]*TextColor, h)
    var i uint
    for i = 0; i < h; i++ {
        grid[i] = store[i*w : (i+1)*w]
    }
    out := Image(grid)
    return &out
}

// convert an image into ASCII!
// watch out, this might be painfully slow...
func Convert(m image.Image, p []*TextColor) *Image {

    c := NewPalette(p)

    // create image of correct size
    bounds := m.Bounds()
    s := bounds.Size()
    img := NewImage(uint(s.X),
        uint(s.Y))

    // dereference for slice manipulation
    grid := *img
    for y := range grid {
        for x := range grid[y] {
            grid[y][x] = c.Convert(m.At(x+bounds.Min.X, y+bounds.Min.Y)).(*TextColor)
        }
    }
    return img
}

// And encoding, oh ho!
// possibly the worst way to write this
func Encode(w io.Writer, m image.Image, s []*TextColor) error {
    // first convert to TextImage
    var local *Image
    if _, ok := m.(*Image); ok {
        local = m.(*Image)
    } else {
        local = Convert(m, s)
    }

    // now convert it to string
    // and use a string writer to write, cause that's easy
    rdr := strings.NewReader(local.String())
    _, err := io.Copy(w, rdr)
    return err // hope this is nil I guess
}

// Some default pixel ratios for fun
// Determine by taking a screenshot of the cursor box in your terminal, and dividing widht/height.
var TextPixelRatio float64 = 6.0 / 14.0
