package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/ricardoecosta/weddingfeed/domain"
	"github.com/ricardoecosta/weddingfeed/imagesaver"
	"github.com/ricardoecosta/weddingfeed/messageservice"
	"mime/multipart"
	"net/http"
	"time"
)

const MAX_UPLOAD_SIZE_BYTES = 16 * 1024 * 1024
const IMAGE_FORM_FIELD = "image"
const SENDER_FORM_FIELD = "sender"
const MESSAGE_FORM_FIELD = "message"

// todo: tests
// todo: log sent errors?
// todo: process in bg, prevent multiple clicks, validate
// todo: see if we can avoid opening the file again
// todo: server timeouts
// todo: catch all /
// todo: avoid encoding message multiple times
// todo: upload to s3
// todo: cursor / pagination?

func HandleMessageCreation(
	writer http.ResponseWriter,
	request *http.Request,
	messageService messageservice.MessageService,
	imageSaver imagesaver.ImageSaver) {

	request.Body = http.MaxBytesReader(writer, request.Body, MAX_UPLOAD_SIZE_BYTES)

	err := request.ParseMultipartForm(MAX_UPLOAD_SIZE_BYTES)
	if err != nil {
		if err == multipart.ErrMessageTooLarge {
			SendError(writer, http.StatusBadRequest, fmt.Sprintf("Message too large maxBytes=%v", MAX_UPLOAD_SIZE_BYTES))
			return
		}

		logrus.Errorf("Unable to parse form: %+v", err.Error())
		SendError(writer, http.StatusBadRequest, "Unable to parse form")
		return
	}

	message := &domain.Message{
		Id:        uuid.New().String(),
		Sender:    request.FormValue(SENDER_FORM_FIELD),
		Message:   request.FormValue(MESSAGE_FORM_FIELD),
		CreatedAt: time.Now().UTC().Unix(),
	}

	uploadedFile, _, err := request.FormFile(IMAGE_FORM_FIELD)
	if err != nil {
		message.ImageAttached = false
		if err != http.ErrMissingFile {
			logrus.Errorf("Unable to read uploaded file: %+v", err)
		}
	} else {
		message.ImageAttached = true
		defer uploadedFile.Close()
	}

	if err = messageService.Upsert(message); err != nil {
		logrus.Errorf("Unable to save message: %+v", err)
		SendError(writer, http.StatusInternalServerError, "Unable to save message")
		return
	}

	if message.ImageAttached {
		imageSaver.ProcessAndSave(message.Id, uploadedFile)
	}

	writer.Header().Set("Location", fmt.Sprintf("/api/messages/%v", message.Id))
	writer.WriteHeader(201)
}
