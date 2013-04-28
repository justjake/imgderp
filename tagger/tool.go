package main

import (
    "os"
    "fmt"
    "github.com/justjake/imgtagger"
    "github.com/justjake/imgtagger/ui"
)

const (
    DefaultLibraryPath = "img_lib.dat"
)

var CurrentLibrary = new(imgtagger.Library)

// pass something to exit the gorgram
var Commands = map[string]ui.Command{
    "quit": ui.QuitCommand,
    "exit": ui.QuitCommand,

    "ls":   ui.ListCommand,
    "list": ui.ListCommand,

    "add" : ui.AddCommand,
}


func main() {
    ui.SetTargetLibrary(CurrentLibrary)
    ui.Run(os.Stdin, os.Stdout, "img> ", Commands)
}
