package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cpustejovsky/event-store/event"
	"strconv"
)

const SnapshotValue string = "SNAPSHOT"

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

// Append takes a context and Envelope and returns an error
// It ensures the Version does not already exist then attempts a PUT operation on the DynamoDB EventStoreTable
func (es *EventStore) Append(ctx context.Context, e *event.Envelope) error {
	//This condition makes sure the sort key Version does not already exist
	input := &dynamodb.PutItemInput{
		TableName: &es.Table,
		Item: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: e.Id},
			//AttributeValueMemberN takes a string value, see https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
			"Version":   &types.AttributeValueMemberN{Value: strconv.Itoa(e.Version)},
			"EventName": &types.AttributeValueMemberS{Value: e.EventName},
			"Event":     &types.AttributeValueMemberB{Value: e.Event},
			"Note":      &types.AttributeValueMemberS{Value: e.Note},
		},
		ConditionExpression: aws.String("attribute_not_exists(Version)"),
	}
	_, err := es.DB.PutItem(ctx, input)
	if err != nil {
		//Using the errors package, the code checks if this is an error specific to the condition being failed and, if so, returns a sentinel error that can be checked
		var errCheck *types.ConditionalCheckFailedException
		if errors.As(err, &errCheck) {
			return &EventAlreadyExistsError{
				ID:      e.Id,
				Version: e.Version,
			}
		}
		return err
	}
	return nil
}

func (es *EventStore) Snapshot(ctx context.Context, agg *event.Envelope) error {
	agg.Note = SnapshotValue
	return es.Append(ctx, agg)
}

// Project takes an id, queries events since the last snapshot, and returns a reconstituted Envelope
func (es *EventStore) Project(ctx context.Context, id string) (*event.Envelope, error) {
	version, err := es.getLastSnapshotVersion(ctx, id)
	if err != nil {
		return nil, err
	}
	params := dynamodb.QueryInput{
		TableName:              aws.String(es.Table),
		KeyConditionExpression: aws.String("Id = :uuid AND Version >= :version"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid":    &types.AttributeValueMemberS{Value: id},
			":version": &types.AttributeValueMemberN{Value: strconv.Itoa(version)},
		},
	}
	events, err := es.query(ctx, &params)
	if err != nil {
		return nil, err
	}
	return event.AggregateEnvelopes(events)
}

// QueryAll takes a context and id and returns a slice of Events and an error
func (es *EventStore) QueryAll(ctx context.Context, id string) ([]event.Envelope, error) {
	params := dynamodb.QueryInput{
		TableName:              aws.String(es.Table),
		KeyConditionExpression: aws.String("Id = :uuid"),
		FilterExpression:       aws.String("Note <> :note"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id},
			":note": &types.AttributeValueMemberS{Value: SnapshotValue},
		},
	}
	return es.query(ctx, &params)
}

func (es *EventStore) getLastSnapshotVersion(ctx context.Context, id string) (int, error) {
	params := dynamodb.QueryInput{
		TableName:              aws.String(es.Table),
		KeyConditionExpression: aws.String("Id = :uuid"),
		FilterExpression:       aws.String("Note = :note"),
		ProjectionExpression:   aws.String("Version"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id},
			":note": &types.AttributeValueMemberS{Value: "SNAPSHOT"},
		},
		Limit:            aws.Int32(1),
		ScanIndexForward: aws.Bool(false),
	}

	events, err := es.query(ctx, &params)
	if err != nil {
		checkErr := &NoEventFoundError{}
		if errors.As(err, &checkErr) {
			return 0, nil
		}
		return 0, err
	}
	return events[0].Version, nil
}

// query takes a context and DynamoDB query parameters and returns a slice of Events and an error
func (es *EventStore) query(ctx context.Context, params *dynamodb.QueryInput) ([]event.Envelope, error) {
	var events []event.Envelope
	// Query paginator provides pagination for queries until there are no more pages for DynamoDB to go through
	// See: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.Pagination.htm
	p := dynamodb.NewQueryPaginator(es.DB, params)
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		// The output is unmarshalled into an Envelope slice which is appended to the events slice
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
