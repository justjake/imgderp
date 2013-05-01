package main

import (
    "flag"
    "os"
    "io"
    "fmt"
    "github.com/justjake/imgtagger/image/resize"
    "github.com/justjake/imgtagger/image/ascii" // lel
    "image"
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
    verbose = flag.Bool("v", false, "Print info about the operation")
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


// working with animated gifs:
    // type GIF struct {
    //     Image     []*image.Paletted // The successive images.
    //     Delay     []int             // The successive delay times, one per frame, in 100ths of a second.
    //     LoopCount int               // The loop count.
    // }
const delayMultiplier = time.Second / 100
func gifAnimate(out *os.File, g *gif.GIF, w, h int, pal []*ascii.TextColor) () {
    // first convert every gif image into a txt image
    resized := make([]*string, len(g.Image))

    first := (&resize.Resizer{g.Image[0], w, h}).ResizeNearestNeighbor().(*image.RGBA)
    prevImg, curImg := first, first
    b := curImg.Bounds()

    for i, frame := range g.Image {
        // overlay imgs
        // gif frames are diffs, this expands to whole images
        curImg = (&resize.Resizer{frame, w, h}).ResizeNearestNeighbor().(*image.RGBA)
        for y := b.Min.Y; y < b.Max.Y; y++ {
            for x := b.Min.X; x < b.Max.X; x++ {
                // copy under non-transparent pixels
                px := curImg.At(x, y)
                if _, _, _, a:= px.RGBA(); a == 0 {
                    curImg.Set(x, y, prevImg.At(x, y))
                }
            }
        }
        // TODO - investigate memory saving in ascii.Convert because of palette duplication
        str := ascii.Convert(curImg, pal).String()
        resized[i] = &str
        prevImg = curImg
    }

    // TODO: deal with registration/transparency color
    // TODO: deal with GIFs that have a frame delay of 0

    // naive playback
    for {
        for i, txt := range resized {
            ts := time.Now().UnixNano()
            showFrame(out, txt)
            used := time.Now().UnixNano() - ts
            time.Sleep(delayMultiplier * time.Duration(g.Delay[i]) - time.Duration(used))
        }
    }
}

// clears the terminal then prints s
func showFrame(out io.Writer, s *string) {
    // clear
    fmt.Fprintln(out, "\033[2J")
    // play
    fmt.Fprintln(out, *s)
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
        decoded, _ := gif.DecodeAll(in)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Animated gif decoding error: %s", err)
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



