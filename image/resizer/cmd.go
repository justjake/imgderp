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


// params
var outputType  = flag.String("ft", "auto", "Output filetype")
var targetWidth = flag.Int("w", 80, "Output width")
var targetHeight = flag.Int("h", 0, "Output height. Will be auto-computed if values is < 1")

func main() {
    // parse args
    flag.Parse()
    args := flag.Args()

    if len(args) < 2 {
        fmt.Println("tool INFILE OUTFILE - resize using nearest neightbor")
        fmt.Println("  supported formats: JPG, PNG, GIF.")
        fmt.Println("  GIFs will be written out as PNGs. Anything can be written to TXT with -ft txt")
        flag.Usage()
        return
    }

    // parse args
    in_name, out_name := args[0], args[1]

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

    resizer := resize.Resize(img)
    if h == 0 {
        h = resizer.HeightForWidth(w)
    }

    // friendly!
    fmt.Printf("Resizing image of type %s to [%d, %d] at '%s'\n", ft, w, h, out_name)

    // resize
    new_img := resizer.NearestNeighbor(w, h)

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
        err = ascii.Encode(out, new_img)
    }

    if err != nil {
        fmt.Println(err)
    }
}



