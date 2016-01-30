package main

import (
    "flag"
    "os"
    "fmt"
    "github.com/justjake/imgtagger/image/resize"
    "github.com/justjake/imgtagger/image/ascii" // lel
    "image"
    "image/color"
    "image/jpeg"
    "image/png"
    "image/gif"
    "runtime/pprof" // cpu profiling
    "log" // TODO: switch most fmt.Fprintf(os.Stderr ... to log.
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
    profile = flag.String("profile", "", "Do animated GIF profiling steps, and write CPU profile to -profile $FILE")
    fdebug  = flag.Int("fdebug", -1, "Debug compiling this frame by printing a preview to stderr, then writing the file to TARGET")
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

func writeOutput(ft string, img image.Image) {
    out, err := getOutputFile()
    if err != nil {
        log.Fatal(err)
    }

    // encode
    switch ft {
        // use lossless PNG for GIFs and other fts.
    default:
        err = png.Encode(out, img)
    case "jpg":
        err = jpeg.Encode(out, img, &jpeg.Options{90})
    case "txt":
        // set up the character encoding
        colors := getColors()

        // write out the text
        err = ascii.Encode(out, img, colors)
    }

    if err != nil {
        log.Fatal(err)
    }
}

func getOutputFile() (*os.File, error) {
    args := flag.Args()
    if len(args) < 2 {
        return os.Stdout, nil
    }

    fn := args[1]
    out, err := os.Create(fn)
    if err != nil {
        return nil, err
    }
    return out, nil
}

func getInputFile() (*os.File, error) {
    args := flag.Args()
    if len(args) == 0 {
        return os.Stdin, nil
    }

    fn := args[0]
    out, err := os.Open(fn)
    if err != nil {
        return nil, err
    }
    return out, nil
}


func main() {
    flag.Parse()

    // FIRST - engage profiling if required
    if *profile != "" {
        proFILE, err := os.Create(*profile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(proFILE)
        defer pprof.StopCPUProfile()
    }

    // parse and handle arguments
    var in, out *os.File
    var out_name string
    args := flag.Args()
    if len(args) >= 2 {
        out_name = args[1]
    }

    if *useTextPixelRatio {
        *pixelRatio = ascii.TextPixelRatio
    }
    w := *targetWidth
    h := *targetHeight

    // open infile
    in, err := getInputFile()
    if err != nil {
        log.Fatal(err)
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
        if *fdebug < 0 {
            gifAnimate(os.Stdout, decoded, r.TargetWidth , r.TargetHeight, getColors())
            return
        } else {
            // debug frame data
            frame := testEncode(decoded, *fdebug, r.TargetWidth, r.TargetHeight, getColors())
            writeOutput("png", frame)
            return
        }
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

    writeOutput(ft, new_img)
}



