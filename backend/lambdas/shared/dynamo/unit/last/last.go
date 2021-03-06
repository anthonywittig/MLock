package last

import (
	"context"
	"fmt"
	"log"
	"mlock/lambdas/shared"
	"mlock/lambdas/shared/dynamo"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Unit struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	PropertyName string `json:"propertyName"`
	CalendarURL  string `json:"calendarUrl"`
	UpdatedBy    string `json:"updatedBy"`
}

const (
	tableName = "Unit"
	itemType  = "1" // Not really using ATM.
)

func Delete(ctx context.Context, name string) error {
	dy, err := dynamo.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err.Error())
	}

	name = strings.TrimSpace(name)

	// TODO: under what circumstances would we want to stop this?

	// No audit trail for deletes. :(

	if _, err = dy.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"type": {S: aws.String(itemType)},
			"name": {S: aws.String(name)},
		},
		TableName: aws.String(tableName),
	}); err != nil {
		return fmt.Errorf("error deleting item: %s", err.Error())
	}

	return nil
}

func Get(ctx context.Context, name string) (Unit, bool, error) {
	dy, err := dynamo.GetClient(ctx)
	if err != nil {
		return Unit{}, false, fmt.Errorf("error getting client: %s", err.Error())
	}

	name = strings.TrimSpace(name)

	result, err := dy.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"type": {S: aws.String(itemType)},
			"name": {S: aws.String(name)},
		},
	})
	if err != nil {
		return Unit{}, false, fmt.Errorf("error getting item: %s", err.Error())
	}
	if result.Item == nil {
		return Unit{}, false, nil
	}

	item := Unit{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return Unit{}, false, fmt.Errorf("error unmarshalling: %s", err.Error())
	}

	return item, true, nil
}

func List(ctx context.Context) ([]Unit, error) {
	dy, err := dynamo.GetClient(ctx)
	if err != nil {
		return []Unit{}, fmt.Errorf("error getting client: %s", err.Error())
	}

	input := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			"type": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(itemType)},
				},
			},
		},
	}

	items := []Unit{}
	for {
		// Get the list of tables
		result, err := dy.QueryWithContext(ctx, input)
		if err != nil {
			return []Unit{}, fmt.Errorf("error calling dynamo: %s", err.Error())
		}

		for _, i := range result.Items {
			item := Unit{}
			if err = dynamodbattribute.UnmarshalMap(i, &item); err != nil {
				return []Unit{}, fmt.Errorf("error unmarshaling: %s", err.Error())
			}
			items = append(items, item)
		}

		input.ExclusiveStartKey = result.LastEvaluatedKey
		if result.LastEvaluatedKey == nil {
			break
		}
	}

	return items, nil
}

func Put(ctx context.Context, oldKey string, item Unit) (Unit, error) {
	dy, err := dynamo.GetClient(ctx)
	if err != nil {
		return Unit{}, fmt.Errorf("error getting client: %s", err.Error())
	}

	item.Name = strings.TrimSpace(item.Name)
	item.PropertyName = strings.TrimSpace(item.PropertyName)
	item.CalendarURL = strings.TrimSpace(item.CalendarURL)

	cd, err := shared.GetContextData(ctx)
	if err != nil {
		return Unit{}, fmt.Errorf("can't get context data: %s", err.Error())
	}

	currentUser := cd.User
	if currentUser == nil {
		return Unit{}, fmt.Errorf("no current user")
	}

	item.Type = itemType
	item.UpdatedBy = currentUser.Email

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return Unit{}, fmt.Errorf("error marshalling map: %s", err.Error())
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dy.PutItemWithContext(ctx, input)
	if err != nil {
		return Unit{}, fmt.Errorf("error putting item: %s", err.Error())
	}

	entity, ok, err := Get(ctx, item.Name)
	if err != nil {
		return Unit{}, err
	}
	if !ok {
		return Unit{}, fmt.Errorf("couldn't find entity after insert")
	}

	if oldKey != "" && oldKey != entity.Name {
		if err := Delete(ctx, oldKey); err != nil {
			return Unit{}, fmt.Errorf("error deleting old item: %s", err.Error())
		}
	}

	return entity, nil
}

func Migrate(ctx context.Context) error {
	exists, err := dynamo.TableExists(ctx, tableName)
	if err != nil {
		return fmt.Errorf("error checking for table: %s", err.Error())
	}
	if exists {
		return nil
	}

	dy, err := dynamo.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err.Error())
	}

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("type"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("name"),
				AttributeType: aws.String("S"),
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("type"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("name"),
				KeyType:       aws.String("RANGE"),
			},
		},
		TableName: aws.String(tableName),
	}

	result, err := dy.CreateTableWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err.Error())
	}

	log.Printf("created table: %s - %+v", tableName, result)

	return nil
}
