package imgtagger

type Tag struct {
    Images []*Image
    Title string
}

type Library struct {
    Tags  []*Tag
    Images  []*Image
}

type Image struct {
    Title string
    Filename string
    Tags []*Tag
}
