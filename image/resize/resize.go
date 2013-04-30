package resize
// Implements Image -> ASCII conversions

import (
    "image"
)

type resizer struct {
    img image.Image
    TargetWidth int
    TargetHeight int
}

// To re-scale in one line:
// new_img := Resizer{img, target_w, target_h}.ResizeNearestNeighbor()
type Resizer resizer



// Pixel ratio is the height (in Px) of a character in your font
// versus it's width. If you use a perfectly square font, or are
// outputting to pixels, 1 is the number.
func (r *resizer) HeightForPixelRatio (rat float64) int {
    H := r.TargetHeight
    h := float64(H) * rat
    return int(h)
}

// Sizing for fit-width
func (r *resizer) HeightForWidth(w int) int {
    W := r.TargetWidth
    H := r.TargetHeight
    scale := float64(w) / float64(W)
    return int(float64(H) * scale)
}

// Sizing for fit-height
func (r *resizer) WidthForHeight(h int) int {
    W := r.TargetWidth
    H := r.TargetHeight
    scale := float64(h) / float64(H)
    return int(float64(W) * scale)
}


// nearest neighbor image scaling
func (r *resizer) ResizeNearestNeighbor() image.Image {
    w, h := r.TargetWidth, r.TargetHeight
    older := r.img
    bounds := older.Bounds()
    oldW, oldH := bounds.Size().X, bounds.Size().Y
    /*bounds.Max.X - bounds.Min.X*/
    /*oldH := bounds.Max.Y - bounds.Min.Y*/

    xFactor := float64(oldW) / float64(w)
    yFactor := float64(oldH) / float64(h)

    newer := image.NewRGBA(image.Rect(0, 0, w, h))

    // iterate over the new image, picking the nearest neighbor from the old image
    for x := 0; x <= w; x++ {
        guess_x := int(xFactor * float64(x)) + bounds.Min.X
        if guess_x > bounds.Max.X {
            guess_x = bounds.Max.X
        }
        for y := 0; y <= w; y++ {
            guess_y := int(yFactor * float64(y)) + bounds.Min.Y
            if guess_y > bounds.Max.Y {
                guess_y = bounds.Max.Y
            }
            newer.Set(x, y, older.At(guess_x, guess_y))
        }
    }
    return newer
}

// if you want to scale just for pixel ratio, or scale down
// based on width
func NewResizer(pic image.Image) *resizer {
    size := pic.Bounds().Size()
    return &resizer{pic, size.X, size.Y}
}
