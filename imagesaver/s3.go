package imagesaver

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
	"github.com/edwvee/exiffix"
	"github.com/pkg/errors"
	"github.com/ricardoecosta/weddingfeed/messageservice"
	"io"
	"os"
	"path"
	"sync"
)

const UPLOAD_JOB_QUEUE_SIZE = 200
const PARALLELISM = 10

type S3ImageService struct {
	options        S3ImageSaverOptions
	client         *s3.S3
	tasks          chan *task
	wg             sync.WaitGroup
	messageService messageservice.MessageService
}

type S3ImageSaverOptions struct {
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type task struct {
	key   string
	image io.ReadSeeker
}

// todo: create s3 bucket dynamically
func NewS3ImageSaver(options S3ImageSaverOptions, messageSvc messageservice.MessageService) (ImageSaver, error) {
	awsCredentials := credentials.NewStaticCredentials(options.AccessKey, options.SecretKey, "")
	config := defaults.Config().WithCredentials(awsCredentials).WithRegion(options.Region)
	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create aws session")
	}
	s3Client := s3.New(awsSession)
	s3ImageService := S3ImageService{
		options:        options,
		client:         s3Client,
		tasks:          make(chan *task, UPLOAD_JOB_QUEUE_SIZE),
		wg:             sync.WaitGroup{},
		messageService: messageSvc,
	}
	return s3ImageService, nil
}

func (imgService S3ImageService) ProcessAndSave(key string, image io.ReadSeeker) {
	imgService.tasks <- &task{image: image, key: key}
}

func (imgService S3ImageService) Start() error {
	for i := 0; i < PARALLELISM; i++ {
		imgService.wg.Add(1)
		go func() {
			defer imgService.wg.Done()
			for task := range imgService.tasks {
				options := imgService.options
				saveResult, err := runTask(imgService.client, options.Region, options.Bucket, task.key, task.image)
				if err != nil {
					logrus.Errorf("Image processing task failed id=%v %+v", task.key, err)
					return
				}

				message, err := imgService.messageService.Get(task.key)
				if err != nil {
					logrus.Errorf("Failed to get message from repository id=%v %+v", task.key, err)
					return
				}

				message.ImageUrl = saveResult.ImageUrl
				message.ImageWidth = saveResult.ImageWidth
				message.ImageHeight = saveResult.ImageHeight
				message.ThumbnailUrl = saveResult.ThumbnailUrl
				message.ThumbnailWidth = saveResult.ThumbnailWidth
				message.ThumbnailHeight = saveResult.ThumbnailHeight

				err = imgService.messageService.Upsert(message)
				if err != nil {
					logrus.Errorf("Failed to update message with image urls id=%v %+v", task.key, err)
					return
				}
			}
		}()
	}
	return nil
}

func (imgService S3ImageService) Stop() error {
	close(imgService.tasks)
	imgService.wg.Wait()
	return nil
}

func runTask(client *s3.S3, region, bucket, key string, reader io.ReadSeeker) (*ImageSaveResult, error) {
	img, _, err := exiffix.Decode(reader)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to decode image key=%v", key))
	}

	var originalWidth, originalHeight int
	tempFilePath := path.Join(os.TempDir(), appendJpegExt(key+ORIGINALS_FOLDER))
	if img.Bounds().Dx() > MAX_ORIGINAL_WIDTH {
		ratio := MAX_ORIGINAL_WIDTH / float64(img.Bounds().Dx())
		height := float64(img.Bounds().Dy()) * ratio
		scaledOriginal := imaging.Resize(img, MAX_ORIGINAL_WIDTH, int(height), imaging.Lanczos)
		if err = imaging.Save(scaledOriginal, tempFilePath, imaging.JPEGQuality(95)); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to save original image to temp file file=%v", tempFilePath))
		}
		originalWidth = scaledOriginal.Bounds().Dx()
		originalHeight = scaledOriginal.Bounds().Dy()
	} else {
		if err = imaging.Save(img, tempFilePath, imaging.JPEGQuality(95)); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to save original image to temp file file=%v", tempFilePath))
		}
		originalWidth = img.Bounds().Dx()
		originalHeight = img.Bounds().Dy()
	}

	tempFile, err := os.Open(tempFilePath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to open temp file file=%v", tempFilePath))
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFilePath)
		logrus.Infof("Deleted temp file id=%v file=%v", key, tempFilePath)
	}()

	thumbnail := imaging.Thumbnail(img, THUMB_WIDTH, THUMB_WIDTH, imaging.Lanczos)
	tempThumbnailPath := path.Join(os.TempDir(), appendJpegExt(key+THUMBNAILS_FOLDER))
	if err = imaging.Save(thumbnail, tempThumbnailPath, imaging.JPEGQuality(95)); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to save thumbnail key=%v", key))
	}
	tempThumbnail, err := os.Open(tempThumbnailPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to open temp file file=%v", tempThumbnailPath))
	}
	defer func() {
		tempThumbnail.Close()
		os.Remove(tempThumbnailPath)
		logrus.Infof("Deleted temp file id=%v file=%v", key, tempThumbnailPath)
	}()

	originalUrl, err := uploadToS3(client, region, bucket, path.Join(ORIGINALS_FOLDER, appendJpegExt(key)), tempFile)
	if err != nil {
		return nil, err
	}
	thumbnailUrl, err := uploadToS3(client, region, bucket, path.Join(THUMBNAILS_FOLDER, appendJpegExt(key)), tempThumbnail)
	if err != nil {
		return nil, err
	}
	return &ImageSaveResult{
		ImageUrl:        originalUrl,
		ImageWidth:      originalWidth,
		ImageHeight:     originalHeight,
		ThumbnailUrl:    thumbnailUrl,
		ThumbnailWidth:  THUMB_WIDTH,
		ThumbnailHeight: THUMB_WIDTH,
	}, nil
}

// Uploads the specified file to S3 with public read access.
func uploadToS3(client *s3.S3, region, bucket, key string, file io.ReadSeeker) (string, error) {
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
		ACL:    aws.String(s3.ObjectCannedACLPublicRead)})
	if err != nil {
		msg := fmt.Sprintf("Unable to upload to s3 bucket key=%v bucket=%v", key, bucket)
		return "", errors.Wrap(err, msg)
	}
	return s3ObjectUrl(region, bucket, key), nil
}

func s3ObjectUrl(region, bucket, key string) string {
	return fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", bucket, region, key)
}

// Appends .jpg extension to the provided key
func appendJpegExt(key string) string {
	return key + ".jpg"
}
