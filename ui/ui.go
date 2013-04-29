package ui

import (
    "fmt"
    "io"
    "bufio"
    "os/exec"
    "strings"
    it "github.com/justjake/imgtagger"
)


type settings struct {
    targetLibrary *it.Library
}


var s = settings{}
func SetTargetLibrary (lib *it.Library) {
    s.targetLibrary = lib
}

type Error string
func (e Error) Error() string {
    return string(e)
}

// Commands
type Command func(io.Writer, []string) error

func QuitCommand(out io.Writer, params []string) error {
    return Error("exit")
}

func ListCommand (o io.Writer, p []string) error {
    fmt.Fprintln(o, "Images:")
    if len(s.targetLibrary.Images) > 0 {
        for _, img := range s.targetLibrary.Images {
            fmt.Fprintln(o, img)
        } 
    } else {
            fmt.Fprintln(o, "no images in library.")
    }
    fmt.Fprintln(o, "Tags:")
    if len(s.targetLibrary.Tags) > 0 {
        for _, tag := range s.targetLibrary.Tags {
            fmt.Fprintln(o, tag)
        }
    } else {
        fmt.Fprintln(o, "no tags in library.")
    }
    return nil
}

// adds an image to the library
func AddCommand (o io.Writer, p []string) error {
    // params
    if len(p) < 2 {
        fmt.Fprintf(o, "Add requires at least 2 parameters.\n")
        return nil
    }
    fn, title, tagNames := p[0], p[1], p[2:]

    // get tags or create them
    tags := make([]*it.Tag, len(tagNames))
    for i, tn := range tagNames {
        tag, ok := s.targetLibrary.Tags[tn]
        if ! ok {
            tag, _ = s.targetLibrary.NewTag(tn, []*it.Image{})
        }
        tags[i] = tag
    }

    _, err := s.targetLibrary.NewImage(fn, title, tags)
    if err != nil {
        fmt.Fprintln(o, err)
        return nil
    }

    fmt.Fprintf(o, "Added new image \"%s\" (%s) to library\n", title, fn)
    return nil
}

// tags image with filename p[0] with tag names p[1:]
// every tag is a #hashtag :P
func TagCommand (o io.Writer, p []string) (err error) {
    if len(p) < 2 {
        fmt.Fprintln(o, "`tag` requires at least two params: a filename indicating an image, and a tag for that image")
        return nil
    }

    if img, ok := s.targetLibrary.Images[p[0]]; ok {
        // img exists, let's find the tags
        for _, tag_name := range p[1:] {

            // all tags are hashtags
            if tag_name[0] != '#' {
                tag_name = "#" + tag_name
            }

            tag := s.targetLibrary.FindTag(tag_name)
            s.targetLibrary.TagImage(img, tag)
        }
    } else {
        fmt.Fprintf(o, "Image with filename \"%s\" not found\n", p[0])
    }
    return nil
}

        



// runs a system command, handling the zany errors and things by doing
// UI stuff.
// why.
func System(in io.Reader, out io.Writer, name string, args []string) {
    cmd := exec.Command(name, args...)
    cmd.Stdin = in
    cmd.Stdout = out
    err := cmd.Run()
    switch err.(type) {
    default:
        fmt.Fprintf(out, "System command error: %s\n", err.Error())
    case *exec.ExitError:
        fmt.Fprintf(out, "Exited %s\n", err.Error())
    case nil:
        return
    }
}

// tokenize a readline
func ReadlineWords(reader *bufio.Reader) (words []string, err error) {
    // read in line
    line, err := reader.ReadString('\n')
    if err != nil {
        return nil, err
    }
    line = strings.TrimSpace(line)

    // split and clean words
    words = strings.Split(line, " ")
    for i, val := range words {
        words[i] = strings.TrimSpace(val)
    }
    return
}


// runs on its own thread for maximum responsiveness
func Run(in io.Reader, out io.Writer, prompt string, commands map[string]Command) {
    // set up buffered input
    reader := bufio.NewReader(in)

    UI: for {
        fmt.Fprintf(out, prompt)

        // read in line
        words, err := ReadlineWords(reader)
        if err != nil {
            fmt.Fprintln(out, "Readline error", err)
            continue UI
        }

        // run system command if starts with !
        if words[0][0] == '!' {
            System(in, out, words[0][1:], words[1:])
            continue UI
        }

        // run command with params
        command, ok := commands[words[0]]
        if ok {
            err = command(out, words[1:])
            if err != nil {
                fmt.Fprintln(out, "done.")
                break UI
            }
        } else {
            fmt.Fprintf(out, "Command not found: %s\n", words[0])
        }
    }
}
