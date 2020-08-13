package domain

import (
	"encoding/json"
)

type Message struct {
	Id              string `json:"id"`
	Sender          string `json:"sender"`
	Message         string `json:"message"`
	ImageAttached   bool   `json:"image_attached"`
	ImageUrl        string `json:"image_url,omitempty"`
	ImageWidth      int    `json:"image_width,omitempty"`
	ImageHeight     int    `json:"image_height,omitempty"`
	ThumbnailUrl    string `json:"thumb_url,omitempty"`
	ThumbnailWidth  int    `json:"thumb_width,omitempty"`
	ThumbnailHeight int    `json:"thumb_height,omitempty"`
	Archived        bool   `json:"archived"`
	CreatedAt       int64  `json:"createdAt"`
}

func (m Message) String() string {
	return toString(m)
}

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m ErrorMessage) String() string {
	return toString(m)
}

func toString(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bytes)
}
