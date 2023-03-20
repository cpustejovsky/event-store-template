package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cpustejovsky/event-store/events"
	"strconv"
)

const SnapshotValue string = "SNAPSHOT"

type AttributeValueMap map[string]types.AttributeValue
type AttributeValueMapList []map[string]types.AttributeValue

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
func (es *EventStore) Append(ctx context.Context, e *events.Envelope) error {
	valueMap := AttributeValueMap{
		"Id": &types.AttributeValueMemberS{Value: e.Id},
		//AttributeValueMemberN takes a string value, see https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
		"Version":   &types.AttributeValueMemberN{Value: strconv.Itoa(e.Version)},
		"EventName": &types.AttributeValueMemberS{Value: e.EventName},
		"Event":     &types.AttributeValueMemberB{Value: e.Event},
	}
	return es.append(ctx, valueMap)
}

func (es *EventStore) Snapshot(ctx context.Context, snapshot *events.Snapshot) error {
	valueMap := AttributeValueMap{
		"Id":            &types.AttributeValueMemberS{Value: snapshot.Id + SnapshotValue},
		"Version":       &types.AttributeValueMemberN{Value: strconv.Itoa(snapshot.Version)},
		"LatestVersion": &types.AttributeValueMemberN{Value: strconv.Itoa(snapshot.LatestVersion)},
		"EventName":     &types.AttributeValueMemberS{Value: snapshot.EventName},
		"Event":         &types.AttributeValueMemberB{Value: snapshot.Event},
	}

	return es.append(ctx, valueMap)
}

// Project takes an id, queries events since the last snapshot, and returns a reconstituted Envelope
func (es *EventStore) Project(ctx context.Context, id string) (*events.Envelope, error) {
	snapshot, err := es.getLatestSnapshot(ctx, id)
	if err != nil {
		checkErr := &NoEventFoundError{}
		if errors.As(err, &checkErr) {
			envelopes, err := es.QueryAll(ctx, id)
			if err != nil {
				return nil, err
			}
			return events.AggregateEnvelopes(envelopes)
		}
		return nil, err
	}
	params := dynamodb.QueryInput{
		TableName:              aws.String(es.Table),
		KeyConditionExpression: aws.String("Id = :uuid AND Version >= :version"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid":    &types.AttributeValueMemberS{Value: id},
			":version": &types.AttributeValueMemberN{Value: strconv.Itoa(snapshot.LatestVersion)},
		},
	}
	ml, err := es.query(ctx, &params)
	if err != nil {
		return nil, err
	}
	var envelopes []events.Envelope
	err = attributevalue.UnmarshalListOfMaps(ml, &envelopes)
	if err != nil {
		return nil, err
	}
	snapshotEnvelope := events.Envelope{
		Id:        snapshot.Id,
		Version:   snapshot.LatestVersion - 1,
		Event:     snapshot.Event,
		EventName: snapshot.EventName,
	}
	envelopes = append(envelopes, snapshotEnvelope)
	return events.AggregateEnvelopes(envelopes)
}

// QueryAll takes a context and id and returns a slice of Events and an error
func (es *EventStore) QueryAll(ctx context.Context, id string) ([]events.Envelope, error) {
	params := dynamodb.QueryInput{
		TableName:              aws.String(es.Table),
		KeyConditionExpression: aws.String("Id = :uuid"),
		FilterExpression:       aws.String("Note <> :note"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id},
			":note": &types.AttributeValueMemberS{Value: SnapshotValue},
		},
	}
	maplist, err := es.query(ctx, &params)
	if err != nil {
		return nil, err
	}
	var events []events.Envelope
	err = attributevalue.UnmarshalListOfMaps(maplist, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (es *EventStore) getLatestSnapshot(ctx context.Context, id string) (*events.Snapshot, error) {
	params := dynamodb.QueryInput{
		TableName:              aws.String(es.Table),
		KeyConditionExpression: aws.String("Id = :uuid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id + SnapshotValue},
		},
		Limit:            aws.Int32(1),
		ScanIndexForward: aws.Bool(false),
	}
	var snapshots []events.Snapshot
	mapList, err := es.query(ctx, &params)
	if err != nil {
		return nil, err
	}
	err = attributevalue.UnmarshalListOfMaps(mapList, &snapshots)

	if err != nil {
		return nil, err
	}
	return &snapshots[0], nil
}

// query takes a context and DynamoDB query parameters and returns a slice of Events and an error
func (es *EventStore) query(ctx context.Context, params *dynamodb.QueryInput) (AttributeValueMapList, error) {
	var maps AttributeValueMapList
	// Query paginator provides pagination for queries until there are no more pages for DynamoDB to go through
	// See: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.Pagination.htm
	p := dynamodb.NewQueryPaginator(es.DB, params)
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		items := out.Items
		maps = append(maps, items...)
	}
	// If the slice is empty, then error is returned
	if len(maps) < 1 {
		return nil, &NoEventFoundError{}
	}
	return maps, nil
}

func (es *EventStore) append(ctx context.Context, valueMap AttributeValueMap) error {
	//This condition makes sure the sort key Version does not already exist
	input := &dynamodb.PutItemInput{
		TableName:           &es.Table,
		Item:                valueMap,
		ConditionExpression: aws.String("attribute_not_exists(Version)"),
	}
	_, err := es.DB.PutItem(ctx, input)
	if err != nil {
		//Using the errors package, the code checks if this is an error specific to the condition being failed and, if so, returns a sentinel error that can be checked
		var errCheck *types.ConditionalCheckFailedException
		if errors.As(err, &errCheck) {
			var out events.Envelope
			err := attributevalue.UnmarshalMap(valueMap, &out)
			if err != nil {
				return err
			}
			return &EventAlreadyExistsError{
				ID:      out.Id,
				Version: out.Version,
			}
		}
		return err
	}
	return nil
}
