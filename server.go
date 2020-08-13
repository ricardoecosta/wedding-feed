package main

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/ricardoecosta/weddingfeed/imagesaver"
	"github.com/ricardoecosta/weddingfeed/messageservice"
	"net/http"
	"os"
)

type Server struct {
	Router *Router
	Config *Config
}

type Config struct {
	Port                        uint32 `json:"port"`
	StaticDir                   string `json:"static_dir"`
	AwsRegion                   string `json:"aws_region"`
	AwsAccessKey                string `json:"aws_access_key"`
	AwsSecretKey                string `json:"aws_secret_key"`
	AwsS3Bucket                 string `json:"aws_s3_bucket"`
	AwsDynamoDBMessageTableName string `json:"aws_dynamo_db_message_table_name"`
	DbDataDir                   string `json:"db_data_dir"`
}

func (c Config) String() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func NewServer(configFile string) (*Server, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Loaded server config file=%v", configFile)

	messageServiceOptions := messageservice.DynamoDBOptions{
		Region:    config.AwsRegion,
		TableName: config.AwsDynamoDBMessageTableName,
		AccessKey: config.AwsAccessKey,
		SecretKey: config.AwsSecretKey,
	}
	messageService, err := messageservice.NewDynamoDB(messageServiceOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create message service")
	}
	logrus.Infoln("Message service created")

	imageSaverOptions := imagesaver.S3ImageSaverOptions{
		Region:    config.AwsRegion,
		Bucket:    config.AwsS3Bucket,
		AccessKey: config.AwsAccessKey,
		SecretKey: config.AwsSecretKey,
	}
	imageSaver, err := imagesaver.NewS3ImageSaver(imageSaverOptions, messageService)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create image saver")
	}
	logrus.Infoln("Image saver created")

	if err = imageSaver.Start(); err != nil {
		return nil, errors.Wrap(err, "Failed to start image saver")
	}
	logrus.Infoln("Image saver started")

	router := createRouter(messageService, imageSaver, config.StaticDir)
	logrus.Infoln("Router created")

	server := Server{
		Router: router,
		Config: config,
	}
	return &server, nil
}

func (s Server) Start() {
	logrus.Fatalf("%+v\n", http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Port), s.Router.muxRouter))
}

func createRouter(
	messageService messageservice.MessageService,
	imageSaver imagesaver.ImageSaver,
	staticDir string) *Router {
	router := NewRouter()

	router.ServeStatic(staticDir, true)
	logrus.Infof("Serving files from '%v'", staticDir)

	router.RegisterRoutes(
		Route{
			Path:   "/",
			Method: "GET",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				http.Redirect(writer, request, "static/index.html", http.StatusTemporaryRedirect)
			},
		},
		Route{
			Path:   "/health",
			Method: "GET",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
			},
		},
		Route{
			Path:   "/api/messages",
			Method: "GET",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				messages, err := messageService.All()
				if err != nil {
					SendError(writer, http.StatusInternalServerError, err.Error())
				}
				json.NewEncoder(writer).Encode(messages)
			},
		},
		Route{
			Path:   "/api/messages",
			Method: "POST",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				HandleMessageCreation(writer, request, messageService, imageSaver)
			},
		},
		Route{
			Path:   "/api/messages/unarchived",
			Method: "GET",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				messages, err := messageService.Unarchived()
				if err != nil {
					SendError(writer, http.StatusInternalServerError, err.Error())
				}
				json.NewEncoder(writer).Encode(messages)
			},
		},
		Route{
			Path:   "/api/messages/{id}/archive",
			Method: "PUT",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				vars := mux.Vars(request)
				id := vars["id"]
				messageService.Archive(id)
			},
		},
		Route{
			Path:   "/api/messages/{id}/unarchive",
			Method: "PUT",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				vars := mux.Vars(request)
				id := vars["id"]
				messageService.Unarchive(id)
			},
		},
		Route{
			Path:   "/wall",
			Method: "GET",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				http.Redirect(writer, request, "static/wall.html", http.StatusTemporaryRedirect)
			},
		})
	return router
}

func loadConfig(configFile string) (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't open config file: %v", configFile)
	}
	var config Config
	json.NewDecoder(file).Decode(&config)
	return &config, nil
}
