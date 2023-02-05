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
	Id                      string
	Version                 int
	CharacterName           string
	CharacterHitPointChange int
	Note                    string
}

type AggregatedEvent struct {
	Id                 string
	CharacterName      string
	CharacterHitPoints int
}

type EventStore struct {
	DB    *dynamodb.Client
	Table string
}

type EventAlreadyExistsError struct {
	ID      string ``
	Version int
}

func (e *EventAlreadyExistsError) Error() string {
	return fmt.Sprintf("event already exists for ID %s and Version %d", e.ID, e.Version)
}

type NoEventFoundError struct{}

func (e *NoEventFoundError) Error() string {
	return "no event found"
}

func New(db *dynamodb.Client, table string) *EventStore {
	return &EventStore{DB: db, Table: table}
}

// Append takes a context and Event and returns an error
// It ensures the Version does not already exist then attempts a PUT operation on the DynamoDB EventStoreTable
func (es *EventStore) Append(ctx context.Context, event *Event) error {
	//This condition makes sure the sort key Version does not already exist
	cond := "attribute_not_exists(Version)"
	input := &dynamodb.PutItemInput{
		TableName: &es.Table,
		Item: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: event.Id},
			//AttributeValueMemberN takes a string value, see https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
			"Version":                 &types.AttributeValueMemberN{Value: strconv.Itoa(event.Version)},
			"CharacterName":           &types.AttributeValueMemberS{Value: event.CharacterName},
			"CharacterHitPointChange": &types.AttributeValueMemberN{Value: strconv.Itoa(event.CharacterHitPointChange)},
			"Note":                    &types.AttributeValueMemberS{Value: event.Note},
		},
		ConditionExpression: &cond,
	}
	_, err := es.DB.PutItem(ctx, input)
	if err != nil {
		//Using the errors package, the code checks if this is an error specific to the condition being failed and, if so, returns a sentinel error that can be checked
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

// Replay takes an id, queries events since the last snapshot, and returns an AggregatedEvent
func (es *EventStore) Replay(ctx context.Context, id string) (*AggregatedEvent, error) {
	events, err := es.QueryFromLastSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	agg := aggregateEvents(events)
	return &agg, nil
}

// ReplayFromBeginning takes an id, queries the entire event store, and returns an AggregatedEvent
func (es *EventStore) ReplayFromBeginning(ctx context.Context, id string) (*AggregatedEvent, error) {
	events, err := es.QueryAll(ctx, id)
	if err != nil {
		return nil, err
	}
	agg := aggregateEvents(events)
	return &agg, nil
}

func (es *EventStore) QueryFromLastSnapshot(ctx context.Context, id string) ([]Event, error) {
	kce := "Id = :uuid"
	fe := "Note = :note"
	var lim int32 = 1
	sfi := false
	params := dynamodb.QueryInput{
		TableName:              &es.Table,
		KeyConditionExpression: &kce,
		FilterExpression:       &fe,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id},
			":note": &types.AttributeValueMemberS{Value: "SNAPSHOT"},
		},
		Limit:            &lim,
		ScanIndexForward: &sfi,
	}
	events, err := es.query(ctx, &params)
	if err != nil {
		checkErr := &NoEventFoundError{}
		if errors.As(err, &checkErr) {
			return es.QueryAll(ctx, id)
		}
		return nil, err
	}
	return es.QuerySinceVersion(ctx, id, events[0].Version)
}

// QueryAll takes a context and id and returns a slice of Events and an error
func (es *EventStore) QueryAll(ctx context.Context, id string) ([]Event, error) {
	kce := "Id = :uuid"
	params := dynamodb.QueryInput{
		TableName:              &es.Table,
		KeyConditionExpression: &kce,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id},
		},
	}
	return es.query(ctx, &params)
}

// QuerySinceVersion takes a context, id, and version that will service as starting point for query
// returns a slice of Events and an error
func (es *EventStore) QuerySinceVersion(ctx context.Context, id string, version int) ([]Event, error) {
	kce := "Id = :uuid AND Version >= :version"
	params := dynamodb.QueryInput{
		TableName:              &es.Table,
		KeyConditionExpression: &kce,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid":    &types.AttributeValueMemberS{Value: id},
			":version": &types.AttributeValueMemberN{Value: strconv.Itoa(version)},
		},
	}
	return es.query(ctx, &params)
}

// query takes a context and DynamoDB query parameters and returns a slice of Events and an error
func (es *EventStore) query(ctx context.Context, params *dynamodb.QueryInput) ([]Event, error) {
	var events []Event
	// Query paginator provides pagination for queries until there are no more pages for DynamoDB to go through
	// See: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.Pagination.htm
	p := dynamodb.NewQueryPaginator(es.DB, params)
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		// The output is unmarshalled into an Event slice which is appended to the events slice
		err = attributevalue.UnmarshalListOfMaps(out.Items, &events)
		if err != nil {
			return nil, err
		}
	}
	// If the slice is empty, then error is returned
	if len(events) < 1 {
		return nil, &NoEventFoundError{}
	}
	return events, nil
}

func aggregateEvents(events []Event) AggregatedEvent {
	var agg AggregatedEvent
	for i, event := range events {
		if i == len(events)-1 {
			agg.Id = event.Id
			agg.CharacterName = event.CharacterName
		}
		agg.CharacterHitPoints += event.CharacterHitPointChange
	}
	return agg
}
