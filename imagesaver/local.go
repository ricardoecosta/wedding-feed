package imagesaver
//
//import (
//	"os"
//	"path"
//	"io"
//	"github.com/edwvee/exiffix"
//	"github.com/disintegration/imaging"
//)
//
//type LocalImageService struct {
//	imageSavingDir string
//}
//
//func NewLocalImageService(imageSavingDir string) ImageSaver {
//	for _, dir := range []string{ORIGINALS, THUMBNAILS, LOWRES} {
//		os.MkdirAll(path.Join(imageSavingDir, dir), os.ModePerm)
//	}
//	return LocalImageService{imageSavingDir}
//}
//
//// Reads the image data and saves with the specified name
//func (imgService LocalImageService) SaveImageAs(name string, reader io.ReadSeeker) (string, error) {
//	img, _, err := exiffix.Decode(reader)
//	if err != nil {
//		return "", err
//	}
//	path := path.Join(imgService.imageSavingDir, ORIGINALS, appendJpegExt(name))
//	if err = imaging.Save(img, path, imaging.JPEGQuality(95)); err != nil {
//		return "", err
//	}
//	return path, nil
//}
//
//// Saves a low resolution version of the provided image
//func (imgService LocalImageService) SaveImageInLowRes(image string) (string, error) {
//	file, err := os.Open(image)
//	img, _, err := exiffix.Decode(file)
//	if err != nil {
//		return "", err
//	}
//	defer file.Close()
//	ratio := 1024 / float64(img.Bounds().Dx())
//	height := float64(img.Bounds().Dy()) * ratio
//	lowres := imaging.Resize(img, 1024, int(height), imaging.Lanczos)
//	filename := filename(image)
//	path := path.Join(imgService.imageSavingDir, LOWRES, appendJpegExt(filename))
//	imaging.Save(lowres, path, imaging.JPEGQuality(95))
//	return path, nil
//}
//
//// Saves a cropped thumbnail of the provided image
//func (imgService LocalImageService) SaveImageThumbnail(image string) (string, error) {
//	file, err := os.Open(image)
//	img, _, err := exiffix.Decode(file)
//	if err != nil {
//		return "", err
//	}
//	defer file.Close()
//	thumbnail := imaging.Thumbnail(img, 512, 512, imaging.Lanczos)
//	filename := filename(image)
//	path := path.Join(imgService.imageSavingDir, THUMBNAILS, appendJpegExt(filename))
//	err = imaging.Save(thumbnail, path, imaging.JPEGQuality(95))
//	if (err != nil) {
//		return "", err
//	}
//	return path, nil
//}