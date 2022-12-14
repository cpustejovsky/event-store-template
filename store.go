package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
)

// Event contains necessary information for our event driven system
type Event struct {
	Id      string
	Version int
}

type EventStore struct {
	DB *dynamodb.Client
}

var EventStoreTable = "event-store"

type EventAlreadyExistsError struct {
	ID      string
	Version int
}

func (e *EventAlreadyExistsError) Error() string {
	return fmt.Sprintf("event already exists for ID %s and Version %d", e.ID, e.Version)
}

func New(db *dynamodb.Client) *EventStore {
	return &EventStore{DB: db}
}

// Append takes a context and Event and returns an error
// It ensures the Version does not already exist then attempts a PUT operation on the DynamoDB EventStoreTable
func (es *EventStore) Append(ctx context.Context, event *Event) error {
	//This condition makes sure the sort key Version does not already exist
	cond := "attribute_not_exists(Version)"
	input := &dynamodb.PutItemInput{
		TableName: &EventStoreTable,
		Item: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: event.Id},
			//AttributeValueMemberN takes a string value, see https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
			"Version": &types.AttributeValueMemberN{Value: strconv.Itoa(event.Version)},
		},
		ConditionExpression: &cond,
	}
	_, err := es.DB.PutItem(ctx, input)
	if err != nil {
		//Using the error package, the code checks if this is an error specific to the condition being failed and, if so, returns a sentinel error that can be checked
		var errCheck *types.ConditionalCheckFailedException
		if errors.As(err, &errCheck) {
			return &EventAlreadyExistsError{
				ID:      event.Id,
				Version: event.Version,
			}
		}
		return err
	}
	return nil
}

// Query takes a context and DynamoDB query parameters and returns a slice of Events and an error
func (es *EventStore) Query(ctx context.Context, queryParams *dynamodb.QueryInput) ([]Event, error) {
	var events []Event
	// Query paginator provides pagination for queries until there are no more pages for DynamoDB to go through
	// See: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.Pagination.htm
	p := dynamodb.NewQueryPaginator(es.DB, queryParams)
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		// The output is unmarshalled into an Event slice which is appended to the events slice
		var pItems []Event
		err = attributevalue.UnmarshalListOfMaps(out.Items, &pItems)
		if err != nil {
			return nil, err
		}

		events = append(events, pItems...)
	}
	// If the slice is empty, then error is returned
	if len(events) < 1 {
		return nil, errors.New("no events found")
	}
	return events, nil
}
