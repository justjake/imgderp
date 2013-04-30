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

const defaultCharSet = " ..o:O128@#"

var charsets =  map[string][]*ascii.TextColor {
    "default": ascii.DefaultSet,
    "box": ascii.UnicodeBoxSet,
}


// params
var outputType  = flag.String("ft", "auto", "Output filetype")
var targetWidth = flag.Int("w", 80, "Output width")
var targetHeight = flag.Int("h", 0, "Output height. Will be auto-computed if values is < 1")
var charSet = flag.String("chars", "" , "Characters, from empty to solid, to use for txt-image conversion")
var pixelRatio = flag.Float64("pxr", 1.0, "pixel ratio of output. Useful for -ft=txt, where fonts are usually taller than they are wide")
var useTextPixelRatio = flag.Bool("tpr", false, fmt.Sprintf("Use text pixel ratio of %f instead of the value of -pxr", ascii.TextPixelRatio))
var charSetName = flag.String("setname", "default", "Charset to use for -ft=txt if -chars is unset. Values: default, box")
func main() {
    // parse args
    flag.Parse()
    args := flag.Args()

    if len(args) < 2 {
        fmt.Println(os.Args[0], "- resize images and possibly transform them to text")
        fmt.Println("USEAGE", os.Args[0], "[options] INFILE OUTFILE")
        fmt.Println("  supported formats: JPG, PNG, GIF.")
        fmt.Println("  GIFs will be written out as PNGs. Anything can be written to TXT with -ft txt")
        flag.Usage()
        return
    }

    // parse args
    in_name, out_name := args[0], args[1]
    if *useTextPixelRatio {
        *pixelRatio = ascii.TextPixelRatio
    }


    w := *targetWidth
    h := *targetHeight

    // open infile
    in, err := os.Open(in_name)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer in.Close()

    // decode image
    img, ft, err := image.Decode(in)
    if err != nil {
        fmt.Println(err)
        return
    }

    // open outfile
    out, err := os.Create(out_name)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer out.Close()

    resizer := resize.NewResizer(img)
    if h == 0 {
        h = resizer.HeightForWidth(w)
    }
    resizer.TargetWidth = w
    resizer.TargetHeight = h
    resizer.TargetHeight = resizer.HeightForPixelRatio(*pixelRatio)

    // friendly!
    fmt.Printf("Resizing image of type %s to [%d, %d] at '%s'\n", ft, w, h, out_name)

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
    case "jpeg":
        err = jpeg.Encode(out, new_img, &jpeg.Options{90})
    case "txt":
        // set up the character encoding
        var colors []*ascii.TextColor
        if *charSet == "" {
            colors = charsets[*charSetName]
        } else {
            colors = ascii.MakeTextColors([]rune(*charSet)...)
        }
        err = ascii.Encode(out, new_img, colors)
    }

    if err != nil {
        fmt.Println(err)
    }
}



