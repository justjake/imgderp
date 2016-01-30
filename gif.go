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


// gif playback steps
func resizeStep(w, h int, in, out chan image.Image) {
    for {
        img := <-in
        out <- (&resize.Resizer{img, w, h, 0, 0}).ResizeNearestNeighbor()
    }
}

func asciiStep(pal []*ascii.TextColor, in chan image.Image, out chan *ascii.Image) {
    i := 0
    for img := range in {
        out <- ascii.ConvertSync(img, pal)
        if *verbose {
            fmt.Fprintf(os.Stderr, "Converted frame %d to ASCII\n", i)
            i++
        }
    }
    if *verbose {
        fmt.Fprintf(os.Stderr, "Done ASCIIing %d frames\n", i)
    }
    close(out)
}

func stringify(img *ascii.Image, index int,  store [][]string) {
    strings := make([]string, len(*img))
    for k := range strings {
        strings[k] = img.StringLine(k)
    }
    store[index] = strings
}

func stringsStep(in chan *ascii.Image, out [][]string, done chan bool) {
    i := 0
    for img := range in {
        stringify(img, i, out)
        if *verbose {
            fmt.Fprintf(os.Stderr, "Finished encoding frame %d (ASYNC)\n", i)
            i++
        }
    }
    done <- true
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
        stringify(textImage, i, frames)

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

func encodeFramesPipeline(g *gif.GIF, w, h int, pal []*ascii.TextColor) (frames [][]string) {
    frames = make([][]string, len(g.Image))

    // set up frame accumulator
    originalBounds := g.Image[0].Bounds()
    compiledImage := image.NewRGBA(originalBounds)
    copyImageOver(compiledImage, g.Image[0])

    // set up channels for processing pipeline
    bufferSize := len(g.Image)
    // cant do resizing on own thread because things will copy over the
    // acucmulator image
    //fullFrames := make(chan image.Image, bufferSize)
    resizedFrames := make(chan image.Image, bufferSize)
    asciiFrames := make(chan *ascii.Image, bufferSize)
    done := make(chan bool)

    // wait for all images to be processed
    var pipelineFinished sync.WaitGroup
    pipelineFinished.Add(bufferSize)

    // start goroutines in processing pipeline
    go asciiStep(pal, resizedFrames, asciiFrames)
    go stringsStep(asciiFrames, frames, done)

    // timestamp!
    ts := time.Now()

    for i, frame := range g.Image {
        // copy the current frame over the previous frame
        copyImageOver(compiledImage, frame)

        // resize then inject into pipeline
        curFrame := (&resize.Resizer{compiledImage, w, h, 0, 0}).ResizeNearestNeighbor()
        resizedFrames <- curFrame

        // print status info if done
        if *verbose {
            fmt.Fprintf(os.Stderr, "Finished shrinking frame %d\n", i)
        }
    }

    if *verbose {
        fmt.Fprintf(os.Stderr, "Waiting for %d frames to render\n", bufferSize)
    }

    // wait for the pipeline
    close(resizedFrames)
    <-done

    if *verbose || *profile != "" {
        fmt.Fprintf(os.Stderr, "Rendered %d frames in %v seconds (%d FPS, ASYNC)\n", bufferSize, time.Since(ts), int(time.Since(ts)) / bufferSize)
    }

    return
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
        encodeFramesPipeline(g, w, h, pal)
    } else {
        frames = encodeFramesSync(g, w, h, pal)
        playback(out, g, frames)
    }

}

// TODO: interlace 60fps
func playbackThreaded(out *os.File, g *gif.GIF, frames [][]string) {
    abort := new(bool)
    *abort = false
    var writeLock sync.Mutex
    for {
        for i, frame := range frames {
            go showFrame(out, frame, abort, writeLock)
            if *verbose {
                fmt.Fprintf(out, " frame %s\n", i)
            }
            time.Sleep(delayMultiplier * time.Duration(g.Delay[i]))
            *abort = true
        }
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


