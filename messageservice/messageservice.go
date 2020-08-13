package messageservice

import (
	"github.com/ricardoecosta/weddingfeed/domain"
)

type MessageService interface {
	Get(id string) (*domain.Message, error)
	Upsert(message *domain.Message) error
	All() ([]*domain.Message, error)
	Unarchived() ([]*domain.Message, error)
	Archive(id string) error
	Unarchive(id string) error
}

type MessageFilter func(message *domain.Message) bool

var AllMessages = func(message *domain.Message) bool {
	return true
}
var UnarchivedMessages = func(message *domain.Message) bool {
	return !message.Archived
}
