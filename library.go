package imgtagger

import (
    "fmt"
    "encoding/gob"
    "os"
)


// tag storage subject to change, access over interface
type Tag struct {
    Title string
    Images map[string]*Image
}

type Library struct {
    Images  map[string]*Image // Image.Filename -> Image
    Tags    map[string]*Tag  // Tag.Title -> Tag
}

type Image struct {
    Filename string
    Title string
    Tags map[string]*Tag
}

// remove cycle
type saveableImage struct {
    Filename string
    Title string
}

type saveableTag struct {
    Title string
    Images map[string]*saveableImage
}

type saveableLibrary struct {
    Images map[string]*saveableImage
    Tags   map[string]*saveableTag
}

// All manipulation occurs on the Library
// for centralized data syncing
func (l *Library) NewTag(t string, imgs []*Image) (tag *Tag, err error) {
    // no duplicate tags
    if _, ok := l.Tags[t]; ok {
        // duplicate tag name
        return nil, fmt.Errorf("Duplicate tag name: %s", t)
    }

    return &Tag{t, make(map[string]*Image, 5)}, nil
}

func (l *Library) NewImage(fn string, t string, tags []*Tag) (img *Image, err error) {
    // no duplicate filenames
    if _, ok := l.Images[fn]; ok {
        // duplicate filename
        return nil, fmt.Errorf("Duplicate filename: %s", fn)
    }

    img = &Image{Filename: fn, Title: t, Tags: make(map[string]*Tag, 5)}
    for _, tag := range tags {
        img.Tags[tag.Title] = tag
        tag.Images[fn] = img
    }

    l.Images[fn] = img

    return
}

// write to disk
func (l *Library) Save(fn string) error {
    file, err := os.Create(fn)
    if err != nil {
        return err
    }
    defer file.Close()

    coder := gob.NewEncoder(file)
    err = coder.Encode(l)
    if err != nil {
        return err
    }
    return nil
}

func NewLibrary() (lib *Library) {
    lib = &Library{
        Images: make(map[string]*Image, 10),
        Tags:   make(map[string]*Tag,   10),
    }
    return
}


// Load a gob-dumped library from a file
func LoadLibrary (fn string) (lib *Library, err error) {
    file, err := os.Open(fn)
    if err != nil {
        return
    }
    defer file.Close()

    g := gob.NewDecoder(file)
    lib = &Library{}
    err = g.Decode(lib)
    return // will return err either way
}

// is this required?
func main() {
    gob.Register(&Tag{})
    gob.Register(&Library{})
    gob.Register(&Image{})
}
