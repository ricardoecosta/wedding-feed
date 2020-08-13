package messageservice

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/pkg/errors"
	"github.com/ricardoecosta/weddingfeed/domain"
	"sort"
)

type DynamoDBOptions struct {
	Region    string
	TableName string
	AccessKey string
	SecretKey string
}

type DynamoDBMessageService struct {
	client    *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDB(options DynamoDBOptions) (MessageService, error) {
	awsCredentials := credentials.NewStaticCredentials(options.AccessKey, options.SecretKey, "")
	config := defaults.Config().WithCredentials(awsCredentials).WithRegion(options.Region)
	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create aws session")
	}
	dynamoDBClient := dynamodb.New(awsSession)

	return &DynamoDBMessageService{
		client:    dynamoDBClient,
		tableName: options.TableName}, nil
}

func (m DynamoDBMessageService) Get(id string) (*domain.Message, error) {
	if len(id) == 0 {
		return nil, errors.New("Message id can't be empty")
	}
	result, err := m.client.Query(&dynamodb.QueryInput{
		TableName: aws.String(m.tableName),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				S: aws.String(id),
			},
		},
		KeyConditionExpression: aws.String("id = :id"),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get message from dynamodb, id=%v", id)
	}
	message := &domain.Message{}
	err = dynamodbattribute.UnmarshalMap(result.Items[0], message)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to unmarshal message from dynamodb, id=%v", id)
	}
	return message, nil
}

func (m DynamoDBMessageService) Upsert(message *domain.Message) error {
	if message == nil {
		return errors.New("Message can't be nil")
	}
	attributeValues, err := dynamodbattribute.MarshalMap(message)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal message in dynamodb, id=%v", message.Id)
	}
	_, err = m.client.PutItem(&dynamodb.PutItemInput{
		Item:      attributeValues,
		TableName: aws.String(m.tableName),
	})
	if err != nil {
		return errors.Wrapf(err, "Failed to put item in dynamodb, id=%v", message.Id)
	}
	return nil
}

func (m DynamoDBMessageService) All() ([]*domain.Message, error) {
	result, err := m.client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(m.tableName),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get all messages")
	}
	messages := make([]*domain.Message, len(result.Items))
	for i, m := range result.Items {
		message := &domain.Message{}
		err = dynamodbattribute.UnmarshalMap(m, message)
		if err != nil {
			logrus.Errorf("Failed to unmarshal message")
		}
		messages[i] = message
	}
	m.sortByMostRecent(messages)
	return messages, nil
}

func (m DynamoDBMessageService) Unarchived() ([]*domain.Message, error) {
	filter := expression.Name("archived").Equal(expression.Value(false))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create dynamodb scan expression")
	}
	result, err := m.client.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(m.tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get unarchived messages")
	}
	messages := make([]*domain.Message, len(result.Items))
	for i, m := range result.Items {
		message := &domain.Message{}
		err = dynamodbattribute.UnmarshalMap(m, message)
		if err != nil {
			logrus.Errorf("Failed to unmarshal message")
		}
		messages[i] = message
	}
	m.sortByMostRecent(messages)
	return messages, nil
}

// todo: use update api
func (m DynamoDBMessageService) Archive(id string) error {
	if len(id) == 0 {
		return errors.New("Message id can't be empty")
	}
	message, err := m.Get(id)
	if err != nil {
		return errors.Errorf("Message not found, id=%v", id)
	}
	message.Archived = true
	err = m.Upsert(message)
	if err != nil {
		errors.Wrapf(err, "Failed to archive message, id=%v", id)
	}
	return nil
}

func (m DynamoDBMessageService) Unarchive(id string) error {
	if len(id) == 0 {
		return errors.New("Message id can't be empty")
	}
	message, err := m.Get(id)
	if err != nil {
		return errors.Errorf("Message not found, id=%v", id)
	}
	message.Archived = false
	err = m.Upsert(message)
	if err != nil {
		errors.Wrapf(err, "Failed to unarchive message, id=%v", id)
	}
	return nil
}

func (m DynamoDBMessageService) sortByMostRecent(messages []*domain.Message) {
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt > messages[j].CreatedAt
	})
}