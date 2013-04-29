package ascii

// implements an image type and color pallete using UTF8 runes
// because.

import (
    "image"
    "image/color"
    "strings"
    "io" // for a silly reason ;)
)

// TextImage type!

// cols[rows[letters]]
type Image [][]rune

func (img *Image) ColorModel() color.Model {
    return TextModel
}

func (img *Image) Bounds() image.Rectangle {
    w := len(*img)
    h := len((*img)[0])

    return image.Rect(0, 0, w, h)
}

func (img *Image) At(x, y int) color.Color {
    return TextColor((*img)[x][y])
}


func (t *Image) String() string {
    grid := [][]rune(*t)
    lines := make([]string, len(grid))
    for i, v := range grid {
        lines[i] = string(v) + "\n"
    }
    return strings.Join(lines, "")
}

// Image creation
func NewImage(w, h uint) *Image {
    size := w * h
    store := make([]rune, size)
    grid := make([][]rune, w)
    var i uint
    for i = 0; i < w; i++ {
        grid[i] = store[i*h:(i+1)*h]
    }
    out := Image(grid)
    return &out
}

// convert an image into ASCII!
// watch out, this might be painfully slow...
func Convert(m image.Image) *Image {

    // check to see if its already ASCII (!)
    if cor, ok := m.(*Image); ok {
        return cor
    }

    // create image of correct size
    bounds := m.Bounds()
    s := bounds.Size()
    img := NewImage(uint(s.X),
                    uint(s.Y))

    // dereference for slice manipulation
    grid := *img
    for x := range grid {
        for y := range grid[x] {
            grid[x][y] = rune(TextModel.Convert(m.At(x + bounds.Min.X, y + bounds.Min.Y)).(TextColor))
        }
    }
    return img
}


// And encoding, oh ho!
// possibly the worst way to write this
func Encode(w io.Writer, m image.Image) error {
    // first convert to TextImage
    var local *Image
    if _, ok := m.(*Image); ok {
        local = m.(*Image)
    } else {
        local = Convert(m)
    }

    // now convert it to string
    // and use a string writer to write, cause that's easy
    rdr := strings.NewReader(local.String())
    _, err := io.Copy(w, rdr)
    return err // hope this is nil I guess
}
