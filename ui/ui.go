package ui

import (
    "fmt"
    "io"
    "bufio"
    "os/exec"
    "strings"
    "github.com/justjake/imgtagger"
)


type settings struct {
    targetLibrary *imgtagger.Library
}

var s = settings{}
func SetTargetLibrary (lib *imgtagger.Library) {
    s.targetLibrary = lib
}

type Error string
func (e Error) Error() string {
    return string(e)
}

// Commands
type Command func(io.Writer, []string) error


// Commands must be passed a certain number of args
func NeedsMoreThan(min int, c Command) Command {
    return func (o io.Writer, p []string) error {
        if l := len(p); l < min {
            fmt.Fprintf(o, "This command needs more than %d args. You passed %d.", min, l)
            return nil
        }
        return c(o, p)
    }
}



var (
    QuitCommand =  func (out io.Writer, params []string) error {
        fmt.Fprintln(out, "Exiting.")
        return Error("exit")
    }

    ListCommand = func(o io.Writer, p []string) error {
        if len(s.targetLibrary.Images) > 0 {
            for _, img := range s.targetLibrary.Images {
                fmt.Fprintln(o, img.Title)
            }
        } else {
            fmt.Fprintln(o, "No images in library.")
        }
        return nil // no errors
    }

    AddCommand = NeedsMoreThan(2, func(o io.Writer, p []string) error {
        return nil
    })
)

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
