package main

import (
	"flag"
	"fmt"
	"github.com/justjake/imgderp/ascii" // lel
	"github.com/justjake/imgderp/derper"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log" // TODO: switch most fmt.Fprintf(os.Stderr ... to log.
	"os"
	"runtime/pprof" // cpu profiling
)

// params
var (
	animate       = flag.Bool("animate", false, "Skip all other functions and just show an animated gif")
	outputType    = flag.String("ft", "auto", "Output filetype")
	targetWidth   = flag.Int("w", 80, "Output width")
	targetHeight  = flag.Int("h", 0, "Output height. Will be auto-computed if values is < 1")
	charSet       = flag.String("chars", "", "Characters, from empty to solid, to use for txt-image conversion")
	pixelRatio    = flag.Float64("pxr", ascii.TextPixelRatio, "pixel ratio of output. Fonts are usually taller than wide, so we default to ")
	charSetName   = flag.String("set", "default", "Charset to use for -ft=txt if -chars is unset. Values: default, box")
	invertCharSet = flag.Bool("invert", false, "Reverse the ordering of the charset. Useful for dark-on-light output")
	verbose       = flag.Bool("v", false, "Print info about the operation and image")
	profile       = flag.String("profile", "", "Do animated GIF profiling steps, and write CPU profile to -profile $FILE")
	fdebug        = flag.Int("fdebug", -1, "Debug compiling this frame by printing a preview to stderr, then writing the file to TARGET")
)

func init() {
	// prints help message
	flag.Usage = func() {
		// char sets
		chars := ""
		for n, s := range derper.Charsets {
			chars += fmt.Sprintf("  \"%s\" - %s\n", n, s)
		}

		fmt.Fprintln(os.Stderr, os.Args[0], "- nearest neighbor image scaling and text output")
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] [INFILE] [OUTFILE]")
		fmt.Fprintln(os.Stderr, "  supported formats: JPG, PNG, GIF.")
		fmt.Fprintln(os.Stderr, "  GIFs will be written out as PNGs. Anything can be written to TXT with -ft txt")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "Built-in character sets:\n"+chars)
		fmt.Fprintf(os.Stderr, `Examples:
  # print funny-cartoon.gif to STDOUT as text
  %s -ft=txt -pxr=0.6 ~/Pictures/funny-cartoon.gif
  # resize screenshot.png to 500px wide, and save as a JPEG
  %s -ft=jpg -w=500 ~/Pictures/screenshot.png ~/public_html/images/screenshot.jpg
`, os.Args[0], os.Args[0])
	}
}

func showPaletteInfo(o *os.File, p color.Palette) {
	for i, clr := range p {
		r, g, b, a := clr.RGBA()
		fmt.Fprintf(o, "[%d] - r: %d, g: %d, b: %d, a: %d\n", i, r, g, b, a)
	}
}

func writeOutput(frames [][]string) {
	out, err := getOutputFile()
	if err != nil {
		log.Fatal(err)
	}

	for _, frame := range frames {
		fmt.Fprintln(out, "")
		for _, line := range frame {
			fmt.Fprintln(out, line)
		}
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

	// SECOND - set derper verbosity
	derper.SetVerbose(*verbose)

	// parse and handle arguments
	var in, out *os.File
	//var out_name string
	//args := flag.Args()
	//if len(args) >= 2 {
	//out_name = args[1]
	//}

	var opts derper.Options = derper.Options{
		TargetWidth:  *targetWidth,
		TargetHeight: *targetHeight,
		PixelRatio:   *pixelRatio,

		CharSet:     *charSet,
		CharSetName: *charSetName,
		Invert:      *invertCharSet,
	}

	// open infile
	in, err := getInputFile()
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

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

	var result [][]string

	// handle gifs since they need to be decoded differently
	if ft == "gif" {
		in.Seek(0, 0)
		gifs, err := gif.DecodeAll(in)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Animated gif decoding error: %s\n", err)
			return
		}

		// display the gif as an amination instead of writing it to the output file
		if *animate {
			gifAnimate(os.Stdout, gifs, &opts)
			return
		}

		result = derper.DerpGif(gifs, &opts)
	} else {
		result = derper.Derp(img, &opts)
	}

	writeOutput(result)

	// TODO: note the operation
	// if *verbose {
	// 	fmt.Fprintf(os.Stderr, "Resizing image of type %s to [%d, %d] ", ft, opts.targetWidth, opts.targetHeight)
	// 	if len(out_name) > 0 {
	// 		fmt.Fprintf(os.Stderr, "at '%s'\n", out_name)
	// 	} else {
	// 		fmt.Fprintf(os.Stderr, "\n")
	// 	}
	// }
}

func gifAnimate(out *os.File, g *gif.GIF, opts *derper.Options) {
	frames := derper.DerpGif(g, opts)

	if *profile != "" {
		return
	}

	derper.Playback(out, g, frames)
}
