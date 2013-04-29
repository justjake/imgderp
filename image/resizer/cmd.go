package main

import (
    "os"
    "fmt"
    "github.com/justjake/imgtagger/image/resize"
    "image"
    "image/jpeg"
    "image/png"
    "strconv"
)

func main() {
    // in and out file args needed
    if len(os.Args) < 4 {
        fmt.Println("tool INFILE OUTFILE WIDTH [HEIGHT] - resize using nearest neightbor")
        fmt.Println("  supported formats: JPG, PNG, GIF.")
        fmt.Println("  GIFs will be written out as PNGs.")
        return
    }

    // parse args
    in_name, out_name := os.Args[1], os.Args[2] 
    w, err := strconv.Atoi(os.Args[3])
    if err != nil {
        fmt.Println(err)
        return
    }

    var h int
    if len(os.Args) > 4 {
        h, err = strconv.Atoi(os.Args[4])
        if err != nil {
            fmt.Println(err)
            return
        }
    }

    // open infile
    in, err := os.Open(in_name)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer in.Close()

    // open outfile
    out, err := os.Create(out_name)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer out.Close()

    // decode image
    img, ft, err := image.Decode(in)
    if err != nil {
        fmt.Println(err)
        return
    }

    resizer := resize.Resize(img)
    if h == 0 {
        h = resizer.HeightForWidth(w)
    }

    // friendly!
    fmt.Printf("Resizing image of type %s to [%d, %d] at '%s'\n", ft, w, h, out_name)

    // resize
    new_img := resizer.NearestNeighbor(w, h)

    // encode
    switch ft {
        // use lossless PNG for GIFs and other fts.
    default:
        err = png.Encode(out, new_img)
    case "jpeg":
        err = jpeg.Encode(out, new_img, &jpeg.Options{90})
    }

    if err != nil {
        fmt.Println(err)
    }
}



