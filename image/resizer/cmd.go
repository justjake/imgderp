package main

import (
    "flag"
    "os"
    "fmt"
    "github.com/justjake/imgtagger/image/resize"
    "github.com/justjake/imgtagger/image/ascii" // lel
    "image"
    "image/jpeg"
    "image/png"
)

var charsets =  map[string][]*ascii.TextColor {
    "default": ascii.DefaultSet,
    "box": ascii.UnicodeBoxSet,
    "alt": ascii.AlternateSet,
    "shade": ascii.UnicodeShadeSet,
}

// params
var (
    outputType  = flag.String("ft", "auto", "Output filetype")
    targetWidth = flag.Int("w", 80, "Output width")
    targetHeight = flag.Int("h", 0, "Output height. Will be auto-computed if values is < 1")
    charSet = flag.String("chars", "" , "Characters, from empty to solid, to use for txt-image conversion")
    pixelRatio = flag.Float64("pxr", 1.0, "pixel ratio of output. Useful for -ft=txt, where fonts are usually taller than they are wide")
    useTextPixelRatio = flag.Bool("tpr", false, fmt.Sprintf("Use text pixel ratio of %f instead of the value of -pxr", ascii.TextPixelRatio))
    charSetName = flag.String("setname", "default", "Charset to use for -ft=txt if -chars is unset. Values: default, box")
    invertCharSet = flag.Bool("invert", false, "Reverse the ordering of the charset. Useful for dark-on-light output")
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
            fmt.Println(err)
            return
        }
    }
    defer in.Close()

    // decode image
    img, ft, err := image.Decode(in)
    if err != nil {
        fmt.Println(err)
        return
    }

    // open outfile
    if out == nil {
        out, err = os.Create(out_name)
        if err != nil {
            fmt.Println(err)
            return
        }
    }
    defer out.Close()

    // setup resizer to the correct height
    resizer := resize.NewResizer(img)
    if h == 0 {
        h = resizer.HeightForWidth(w)
    }
    resizer.TargetWidth = w
    resizer.TargetHeight = h
    resizer.TargetHeight = resizer.HeightForPixelRatio(*pixelRatio)

    // note the operation
    fmt.Fprintf(os.Stderr, "Resizing image of type %s to [%d, %d] ", ft, w, h)
    if len(out_name) > 0 {
        fmt.Fprintf(os.Stderr, "at '%s'\n", out_name)
    } else {
        fmt.Fprintf(os.Stderr, "\n")
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
        } else {
            fmt.Println("derp. not inverting")
        }

        // write out the text
        err = ascii.Encode(out, new_img, colors)
    }

    if err != nil {
        fmt.Println(err)
    }
}



