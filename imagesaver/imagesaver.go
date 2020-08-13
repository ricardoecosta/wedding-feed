package imagesaver

import (
	"io"
)

const ORIGINALS_FOLDER = "originals"
const THUMBNAILS_FOLDER = "thumbnails"

const MAX_ORIGINAL_WIDTH = 1080
const THUMB_WIDTH = 256

type ImageSaver interface {
	Start() error
	Stop() error
	ProcessAndSave(key string, image io.ReadSeeker)
}

type ImageSaveResult struct {
	ImageUrl        string
	ImageWidth      int
	ImageHeight     int
	ThumbnailUrl    string
	ThumbnailWidth  int
	ThumbnailHeight int
}
