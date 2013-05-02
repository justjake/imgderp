package ascii

// implements an image type and color pallete using UTF8 runes
// because.

import (
    "image"
    "image/color"
    "io" // for a silly reason ;)
    "strings"
    "sync"
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

func (t *Image) StringLine (y int) string {
    line := (*t)[y]
    out := make([]rune, len(line))
    for i := range line {
        out[i] = line[i].Rune
    }
    return string(out)
}

func (t *Image) String() string {
    grid := *t
    lines := make([]string, len(grid))
    for y :=  range grid {
        lines[y] = t.StringLine(y) + "\n"
    }
    return strings.Join(lines, "")
}

// Image creation
func NewImage(w, h uint) *Image {
    //size := w * h
    // store := make([]*TextColor, size)
    grid := make([][]*TextColor, h)
    var i uint
    for i = 0; i < h; i++ {
        //grid[i] = store[i*w : (i+1)*w]
        grid[i] = make([]*TextColor, w)
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
    var wg sync.WaitGroup
    for y := range grid {
        wg.Add(1)
        go func(r []*TextColor, y int) {
            for x := range r {
                r[x] = c.Convert(m.At(x+bounds.Min.X, y+bounds.Min.Y)).(*TextColor)
            }
            wg.Done()
        }(grid[y], y)
    }
    wg.Wait()
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
