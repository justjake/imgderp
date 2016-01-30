package main
// This file handles animated GIF playback

import (
    "sync"
    "os"
    "io"
    "fmt"
    "github.com/justjake/imgderp/resize"
    "github.com/justjake/imgderp/ascii" // lel
    "image"
    "image/color"
    "image/gif"
    "time" // gif playback
)

// Gif delay multiplier, in nanoseconds
const delayMultiplier = time.Second / 100

// for debugging
// prints the GIF union, then returns the frame pixmap
func testEncode(g *gif.GIF, idx, w, h int, pal []*ascii.TextColor) image.Image {
    // set up frame accumulator
    originalBounds := g.Image[0].Bounds()
    origSizer := resize.Resizer{g.Image[0], w, h, 0, 0}
    compiledImage := origSizer.ResizeNearestNeighbor()
    bg := image.NewUniform(color.RGBA{255, 255, 0, 255})

    compiledImage = image.NewRGBA(compiledImage.Bounds())
    frame := shrinkFrameToCorrectSize(g.Image[idx], w, h, &originalBounds)
    copyImageOver(compiledImage, bg)
    copyImageOver(compiledImage, frame)

    fmt.Fprintln(os.Stderr, ascii.Convert(frame, pal).String())
    fmt.Fprintln(os.Stderr, frame.Bounds())

    return compiledImage
}

// copy a semi-trasparent image over another image in-place
func copyImageOver (base *image.RGBA, newer image.Image)  {
    // points outside of the base bounds will not be copied
    b := base.Bounds().Intersect(newer.Bounds())
    for y := b.Min.Y; y < b.Max.Y; y++ {
        for x := b.Min.X; x < b.Max.X; x++ {
            // copy over non-transparent pixels
        if *verbose {
            //fmt.Fprintf(os.Stderr, "X = %d\n", x)
        }
            px := newer.At(x, y)
            if _, _, _, a:= px.RGBA(); a != 0 {
                base.Set(x, y, px)
            }
        }
    }
}

func stringify(img *ascii.Image) []string {
    strings := make([]string, len(*img))
    for k := range strings {
        strings[k] = img.StringLine(k)
    }
    return strings
}

func encodeFramesSync(g *gif.GIF, w, h int, pal []*ascii.TextColor) (frames [][]string) {
    frames = make([][]string, len(g.Image))

    // set up frame accumulator
    originalBounds := g.Image[0].Bounds()
    origSizer := resize.Resizer{g.Image[0], w, h, 0, 0}
    compiledImage := origSizer.ResizeNearestNeighbor()

    // timestamp!
    ts := time.Now()

    for i, frame := range g.Image {
        // resize the current frame partial
        smallFrame := shrinkFrameToCorrectSize(frame, w, h, &originalBounds)
        if *verbose {
            fmt.Fprintf(os.Stdout, "Shrank frame from bounds %s to new smaller bounds %s\n", frame.Bounds(), smallFrame.Bounds())
        }

        // copy the current frame over the previous frame
        copyImageOver(compiledImage, smallFrame)

        // convert to ascii
        textImage := ascii.ConvertSync(compiledImage, pal)

        // convert to []string and store
        frames[i] = stringify(textImage)

        // print status info if done
        if *verbose {
            fmt.Fprintf(os.Stderr, "Finished encoding frame %d (SYNC)\n", i)
        }
    }

    if *verbose || *profile != "" {
        fmt.Fprintf(os.Stderr, "Rendered %d frames in %v seconds (%d FPS, SYNC)\n", len(g.Image), time.Since(ts), float64(time.Since(ts)) / float64(len(g.Image)))
    }

    return
}

// shrink a frame, keeping its size proportional to the starting image bounds, firstFrame
func shrinkFrameToCorrectSize(frame image.Image, w, h int, firstBounds *image.Rectangle) *image.RGBA {
    frameBounds := frame.Bounds()
    scaleW := float64(w) / float64(firstBounds.Dx())
    scaleH := float64(h) / float64(firstBounds.Dy())

    innerW := int(scaleW * float64(frameBounds.Dx()))
    innerH := int(scaleH * float64(frameBounds.Dy()))

    x := int(float64(frameBounds.Min.X) * scaleW)
    y := int(float64(frameBounds.Min.Y) * scaleH)

    s := resize.Resizer{frame, innerW, innerH, x, y}
    return s.ResizeNearestNeighbor()
}


// working with animated gifs:
    // type GIF struct {
    //     Image     []*image.Paletted // The successive images.
    //     Delay     []int             // The successive delay times, one per frame, in 100ths of a second.
    //     LoopCount int               // The loop count.
    // }

func gifAnimate(out *os.File, g *gif.GIF, w, h int, pal []*ascii.TextColor) () {
    var frames [][]string

    if *profile != "" {
        encodeFramesSync(g, w, h, pal)
    } else {
        frames = encodeFramesSync(g, w, h, pal)
        playback(out, g, frames)
    }

}

func playback(out *os.File, g *gif.GIF, frames [][]string) {
    abort := new(bool)
    *abort = false
    var writeLock sync.Mutex
    for {
        for i, frame := range frames {
            showFrame(out, frame, abort, writeLock)
            if *verbose {
                fmt.Fprintf(out, " frame %d\n", i)
                // showPaletteInfo(out, g.Image[i].Palette)
                fmt.Fprintln(out, g.Image[i].Bounds())
            }
            if delay := g.Delay[i]; delay == 0 {
                time.Sleep(delayMultiplier)
            } else {
                time.Sleep(delayMultiplier * time.Duration(delay))
            }

            *abort = true
        }
    }
}

// clears the terminal then prints s
func showFrame(out io.Writer, f []string, quit *bool, m sync.Mutex) {
    // don't output when another frame is in the pipeline
    m.Lock()
    defer m.Unlock()
    *quit = false

    // clear screen
    fmt.Fprintln(out, "\033[2J")
    // print lines
    for _, line := range f {
        switch *quit {
        case true:
            // out of time. die
            return
        default:
            fmt.Fprintln(out, line)
        }
    }
}
