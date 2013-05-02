package main

import (
    "sync"
    "flag"
    "os"
    "io"
    "fmt"
    "github.com/justjake/imgtagger/image/resize"
    "github.com/justjake/imgtagger/image/ascii" // lel
    "image"
    "image/color"
    "image/jpeg"
    "image/png"
    "image/gif"
    "time" // gif playback
)

var charsets =  map[string][]*ascii.TextColor {
    "default": ascii.DefaultSet,
    "box": ascii.UnicodeBoxSet,
    "alt": ascii.AlternateSet,
    "shade": ascii.UnicodeShadeSet,
    "new": ascii.SciSet,
}

// params
var (
    animate = flag.Bool("animate", false, "Skip all other functions and just show an animated gif")
    outputType  = flag.String("ft", "auto", "Output filetype")
    targetWidth = flag.Int("w", 80, "Output width")
    targetHeight = flag.Int("h", 0, "Output height. Will be auto-computed if values is < 1")
    charSet = flag.String("chars", "" , "Characters, from empty to solid, to use for txt-image conversion")
    pixelRatio = flag.Float64("pxr", 1.0, "pixel ratio of output. Useful for -ft=txt, where fonts are usually taller than they are wide")
    useTextPixelRatio = flag.Bool("tpr", false, fmt.Sprintf("Use text pixel ratio of %f instead of the value of -pxr", ascii.TextPixelRatio))
    charSetName = flag.String("set", "default", "Charset to use for -ft=txt if -chars is unset. Values: default, box")
    invertCharSet = flag.Bool("invert", false, "Reverse the ordering of the charset. Useful for dark-on-light output")
    verbose = flag.Bool("v", false, "Print info about the operation and image")
)

func init() {
// prints help message
    flag.Usage = func() {
        // char sets
        chars := ""
        for n, s := range charsets {
            chars += fmt.Sprintf("  \"%s\" - %s\n", n, s)
        }

        fmt.Fprintln(os.Stderr, os.Args[0], "- nearest neighbor image scaling and text output")
        fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] [INFILE] [OUTFILE]")
        fmt.Fprintln(os.Stderr, "  supported formats: JPG, PNG, GIF.")
        fmt.Fprintln(os.Stderr, "  GIFs will be written out as PNGs. Anything can be written to TXT with -ft txt")
        fmt.Fprintln(os.Stderr, "Options:")
        flag.PrintDefaults()
        fmt.Fprintln(os.Stderr, "Built-in character sets:\n" +  chars)
        fmt.Fprintf(os.Stderr, `Examples:
  # print funny-cartoon.gif to STDOUT as text
  %s -ft=txt -pxr=0.6 ~/Pictures/funny-cartoon.gif
  # resize screenshot.png to 500px wide, and save as a JPEG
  %s -ft=jpg -w=500 ~/Pictures/screenshot.png ~/public_html/images/screenshot.jpg
`, os.Args[0], os.Args[0])
    }
}


// copy a semi-trasparent image over another image in-place
func copyImageOver (base *image.RGBA, newer image.Image)  {
    // points outside of the base bounds will not be copied
    b := base.Bounds().Intersect(newer.Bounds())
    for y := b.Min.Y; y < b.Max.Y; y++ {
        for x := b.Min.X; x < b.Max.X; x++ {
            // copy over non-transparent pixels
            px := newer.At(x, y)
            if _, _, _, a:= px.RGBA(); a != 0 {
                base.Set(x, y, px)
            }
        }
    }
}

// working with animated gifs:
    // type GIF struct {
    //     Image     []*image.Paletted // The successive images.
    //     Delay     []int             // The successive delay times, one per frame, in 100ths of a second.
    //     LoopCount int               // The loop count.
    // }
const delayMultiplier = time.Second / 100
func gifAnimate(out *os.File, g *gif.GIF, w, h int, pal []*ascii.TextColor) () {
    // first convert every gif image into a txt image
    resized := make([][]string, len(g.Image))

    originalBounds := g.Image[0].Bounds()
    originalSize := originalBounds.Size()

    compiledImage := image.NewRGBA(originalBounds)
    copyImageOver(compiledImage, g.Image[0])

    // TODO: copy forward the first frame, overlaying the later frames
    // copy under is a gross hack and obvs. not working
    for i, frame := range g.Image {
        // gif frames are diffs, this expands to whole images
        // compute intended w/h for this one
        // because we'll get streatch sub regiions otherwise
        curFrameOrigSize := frame.Bounds().Size()
        scaleX := float64(curFrameOrigSize.X) / float64(originalSize.X)
        scaleY := float64(curFrameOrigSize.Y) / float64(originalSize.Y)

        // computed target w/h for this frame
        W := int(float64(w) * scaleX)
        H := int(float64(h) * scaleY)

        // copy the current frame over the previous frame
        curFrame := (&resize.Resizer{frame, W, H}).ResizeNearestNeighbor()
        copyImageOver(compiledImage, curFrame)

        // convert image to string in seperate thread
        textImage := ascii.Convert(compiledImage, pal)
        // convert ascii.Image to []string
        strings := make([]string, len(*textImage))
        for k := range strings {
            strings[k] = textImage.StringLine(k)
        }
        // save conversion
        resized[i] = strings
    }

    // TODO: deal with registration/transparency color
    // TODO: deal with GIFs that have a frame delay of 0

    // threaded playback
    playback(out, g, resized)
    playbackThreaded(out, g, resized)
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

func getColors() []*ascii.TextColor {
    var colors []*ascii.TextColor

    // preset vs provided
    if *charSet == "" {
        colors = charsets[*charSetName]
    } else {
        colors = ascii.MakeTextColors([]rune(*charSet)...)
    }

    // flip charset for dark-on-light?
    if *invertCharSet {
        colors = ascii.Reverse(colors)
    }
    return colors
}

func userResizer (img image.Image, w, h int) *resize.Resizer {
    resizer := resize.NewResizer(img)
    if w == 0 && h == 0 {
        w = 80
    }
    if h == 0 {
        h = resizer.HeightForWidth(w)
    }
    if w == 0 {
        w = resizer.WidthForHeight(h)
    }
    resizer.TargetWidth = w
    resizer.TargetHeight = h
    resizer.TargetHeight = resizer.HeightForPixelRatio(*pixelRatio)
    return resizer
}

func showPaletteInfo(o *os.File, p color.Palette) {
    for i, clr := range p {
        r,g,b,a := clr.RGBA()
        fmt.Fprintf(o, "[%d] - r: %d, g: %d, b: %d, a: %d\n", i, r, g, b, a)
    }
}



func main() {
    // parse args
    flag.Parse()
    args := flag.Args()

    var (
        in  *os.File 
        out *os.File
        err error

        in_name string
        out_name string
    )

    switch len(args) {
    case 0:
        // stdin -> stdout
        in = os.Stdin
        out = os.Stdout
    case 1:
        out = os.Stdout
        in_name = args[0]
    case 2:
        in_name = args[0]
        out_name = args[1]
    }


    // parse args
    if *useTextPixelRatio {
        *pixelRatio = ascii.TextPixelRatio
    }
    w := *targetWidth
    h := *targetHeight

    // open infile
    if in == nil {
        in, err = os.Open(in_name)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            return
        }
    }
    defer in.Close()


    // CHOO CHOO STOP HERE AND ANIMATE GIFS IF THATS WHAT WE DO
    if *animate {
        decoded, err := gif.DecodeAll(in)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Animated gif decoding error: %s\n", err)
            return
        }

        // determine image size
        r := userResizer(decoded.Image[0], w, h)
        gifAnimate(os.Stdout, decoded, r.TargetWidth , r.TargetHeight, getColors())
        return
    }

    // decode image
    img, ft, err := image.Decode(in)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
    }

    // verbose -- print image information
    if *verbose {
        paletted, ok := img.(*image.Paletted)
        if ok {
            showPaletteInfo(out, paletted.Palette)
        }
    }

    // open outfile
    if out == nil {
        out, err = os.Create(out_name)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            return
        }
    }
    defer out.Close()

    // setup resizer to the correct height
    resizer := userResizer(img, w, h)

    // note the operation
    if *verbose {
        fmt.Fprintf(os.Stderr, "Resizing image of type %s to [%d, %d] ", ft, w, h)
        if len(out_name) > 0 {
            fmt.Fprintf(os.Stderr, "at '%s'\n", out_name)
        } else {
            fmt.Fprintf(os.Stderr, "\n")
        }
    }

    // resize
    new_img := resizer.ResizeNearestNeighbor()


    // output format selection
    if *outputType != "auto" {
        ft = *outputType
    }

    // encode
    switch ft {
        // use lossless PNG for GIFs and other fts.
    default:
        err = png.Encode(out, new_img)
    case "jpg":
        err = jpeg.Encode(out, new_img, &jpeg.Options{90})
    case "txt":
        // set up the character encoding
        colors := getColors()

        // write out the text
        err = ascii.Encode(out, new_img, colors)
    }

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
}



