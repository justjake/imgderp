package resize
// Implements Image -> ASCII conversions

import (
    "image"
)

type resizer struct {
    img image.Image
}

func (r *resizer) HeightForWidth(w int) int {
    older := r.img
    bounds := older.Bounds()
    W := bounds.Max.X - bounds.Min.X
    H := bounds.Max.Y - bounds.Min.Y
    scale := float64(w) / float64(W)
    return int(float64(H) * scale)
}

func (r *resizer) NearestNeighbor(w, h int) image.Image {
    older := r.img
    bounds := older.Bounds()
    oldW := bounds.Max.X - bounds.Min.X
    oldH := bounds.Max.Y - bounds.Min.Y

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

func Resize(pic image.Image) *resizer {
    return &resizer{pic}
}
