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

// low-level tag creation. FindTag should usually be used.
func (l *Library) NewTag(t string, imgs []*Image) (tag *Tag, err error) {
    // no duplicate tags
    if _, ok := l.Tags[t]; ok {
        // duplicate tag name
        return nil, fmt.Errorf("Duplicate tag name: %s", t)
    }

    img_space := 5
    if len(imgs) > 0 {
        img_space = len(imgs)
    }

    tag = &Tag{t, make(map[string]*Image, img_space)}
    l.Tags[t] = tag

    // tag optional images
    if len(imgs) > 0 {
        for _, img := range imgs {
            l.TagImage(img, tag)
        }
    }

    return tag, nil
}

// find or create the tag with a given title
func (l *Library) FindTag(t string) *Tag {
    tag, ok := l.Tags[t]
    if ! ok {
        tag, _ = l.NewTag(t, []*Image{})
    }
    return tag
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

func (l *Library) TagImage(img *Image, tag *Tag) {
    img.Tags[tag.Title] = tag
    tag.Images[img.Filename] = img
}

// write to disk
func (l *Library) Save(fn string) error {

    // convert to cycle-free data structure
    saveable := &saveableLibrary{
        Images: make(map[string]*saveableImage, len(l.Images)),
        Tags:   make(map[string]*saveableTag,   len(l.Tags)),
    }
    for _, img := range l.Images {
        saveable.Images[img.Filename] = &saveableImage{img.Filename, img.Title}
    }
    for _, tag := range l.Tags {
        st := &saveableTag{
            Title: tag.Title,
            Images: make(map[string]*saveableImage, len(tag.Images)),
        }
        // don't allocate images again
        for fn, _ := range tag.Images {
            st.Images[fn] = saveable.Images[fn]
        }
        saveable.Tags[tag.Title] = st
    }
        
    // file writing
    file, err := os.Create(fn)
    if err != nil {
        return err
    }
    defer file.Close()

    coder := gob.NewEncoder(file)
    err = coder.Encode(saveable)
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
    saveable := &saveableLibrary{}
    err = g.Decode(saveable)
    if err != nil {
        return nil, err
    }

    // decode library
    lib = &Library{
        Images: make(map[string]*Image, len(saveable.Images) + 10),
        Tags:   make(map[string]*Tag, len(saveable.Tags) + 10),
    }
    for _, si := range saveable.Images {
        lib.NewImage(si.Filename, si.Title, []*Tag{})
    }
    for tn, st := range saveable.Tags {
        // can't error because duplicate tags in hash are impossible
        tag, _ := lib.NewTag(tn, []*Image{})
        for img_fn := range st.Images {
            lib.TagImage(lib.Images[img_fn], tag)
        }
    }

    return // will return err either way
}
