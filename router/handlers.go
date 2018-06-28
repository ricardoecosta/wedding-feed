package router

import (
	"net/http"
	"log"
	"mime"
	"github.com/google/uuid"
	"os"
	"path"
	"io"
	"github.com/edwvee/exiffix"
	"github.com/disintegration/imaging"
	"image/png"
	"github.com/ricardoecosta/weddingfeed/domain"
	"encoding/json"
)

const MAX_UPLOAD_SIZE_BYTES = 16 * 1024 * 1024
const UPLOAD_DIRECTORY = "uploads"
const PHOTO_FORM_FIELD = "photo"
const SENDER_FORM_FIELD = "sender"
const MESSAGE_FORM_FIELD = "message"

// todo: log sent errors?
// todo: full hacked function, refactor
// todo: process in bg, prevent multiple clicks, validate
func createMessage(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, MAX_UPLOAD_SIZE_BYTES)

	err := request.ParseMultipartForm(MAX_UPLOAD_SIZE_BYTES)
	if err != nil {
		sendError(writer, http.StatusBadRequest, "Unable to parse multipart form"+err.Error())
		return
	}

	message := &domain.Message{
		Sender:  request.FormValue(SENDER_FORM_FIELD),
		Message: request.FormValue(MESSAGE_FORM_FIELD),
	}

	uploadedFile, uploadedFileHeader, err := request.FormFile(PHOTO_FORM_FIELD)
	if err != nil {
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			sendError(writer, http.StatusInternalServerError, "Unable marshal message"+err.Error())
			return
		}
		writer.Write(jsonMessage)
		return
	}

	defer uploadedFile.Close()

	fileContentType := uploadedFileHeader.Header.Get("Content-Type")
	extensions, err := mime.ExtensionsByType(fileContentType)
	if err != nil {
		sendError(writer, http.StatusBadRequest, "Unable infer file extension from content type, mimeType="+fileContentType)
		return
	}

	uniquePhotoName := uuid.New().String()
	photoExtension := extensions[0]

	localPhotoPath := path.Join(UPLOAD_DIRECTORY, uniquePhotoName+photoExtension)
	localPhotoFile, err := os.OpenFile(localPhotoPath, os.O_WRONLY|os.O_CREATE, 0666)

	io.Copy(localPhotoFile, uploadedFile)
	defer localPhotoFile.Close()

	log.Println("created file, file=" + localPhotoPath)

	// todo: see if we can avoid opening the file again
	localPhotoFile, err = os.Open(localPhotoFile.Name())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	defer localPhotoFile.Close()

	img, _, err := exiffix.Decode(localPhotoFile)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	//dstImage128 := imaging.Resize(img, 512, 0, imaging.Lanczos)
	dstImage128 := imaging.Thumbnail(img, 512, 512, imaging.Lanczos)

	log.Println(uniquePhotoName + "128" + photoExtension)

	err = imaging.Save(dstImage128, path.Join("uploads", uniquePhotoName+"128"+photoExtension), imaging.PNGCompressionLevel(png.BestSpeed))
	if (err != nil) {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

}

func getMessages(writer http.ResponseWriter, request *http.Request) {

}

func updateMessage(writer http.ResponseWriter, request *http.Request) {

}

func getMessage(writer http.ResponseWriter, request *http.Request) {

}

func deleteMessage(writer http.ResponseWriter, request *http.Request) {

}

func sendError(writer http.ResponseWriter, code int, message string) {
	error := &domain.Error{Code: code, Message: message}
	jsonError, err := json.Marshal(error)
	if err != nil {
		http.Error(writer, "Unable to marshal error", http.StatusInternalServerError)
	}
	http.Error(writer, string(jsonError), code)
}
