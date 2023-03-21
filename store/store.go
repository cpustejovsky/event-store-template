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

type EventStore interface {
	Append(context.Context, *events.Envelope) error
	Snapshot(context.Context, *events.Snapshot) error
	Project(context.Context, string) (*events.Envelope, error)
	QueryLatestVersion(context.Context, string) (int, error)
	QueryAll(context.Context, string) ([]events.Envelope, error)
}

type DynamoDBEventStore struct {
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

func DynamoDB(db *dynamodb.Client, table string) *DynamoDBEventStore {
	return &DynamoDBEventStore{DB: db, Table: table}
}

// Append takes a context and Envelope and returns an error
// It ensures the Version does not already exist then attempts a PUT operation on the DynamoDB EventStoreTable
func (d *DynamoDBEventStore) Append(ctx context.Context, e *events.Envelope) error {
	valueMap := AttributeValueMap{
		"Id": &types.AttributeValueMemberS{Value: e.Id},
		//AttributeValueMemberN takes a string value, see https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
		"Version":   &types.AttributeValueMemberN{Value: strconv.Itoa(e.Version)},
		"EventName": &types.AttributeValueMemberS{Value: e.EventName},
		"Event":     &types.AttributeValueMemberB{Value: e.Event},
	}
	return d.append(ctx, valueMap)
}

func (d *DynamoDBEventStore) Snapshot(ctx context.Context, snapshot *events.Snapshot) error {
	valueMap := AttributeValueMap{
		"Id":            &types.AttributeValueMemberS{Value: snapshot.Id + SnapshotValue},
		"Version":       &types.AttributeValueMemberN{Value: strconv.Itoa(snapshot.Version)},
		"LatestVersion": &types.AttributeValueMemberN{Value: strconv.Itoa(snapshot.LatestVersion)},
		"EventName":     &types.AttributeValueMemberS{Value: snapshot.EventName},
		"Event":         &types.AttributeValueMemberB{Value: snapshot.Event},
	}

	return d.append(ctx, valueMap)
}

// Project takes an id, queries events since the last snapshot, and returns a reconstituted Envelope
func (d *DynamoDBEventStore) Project(ctx context.Context, id string) (*events.Envelope, error) {
	snapshot, err := d.getLatestSnapshot(ctx, id)
	if err != nil {
		checkErr := &NoEventFoundError{}
		if errors.As(err, &checkErr) {
			envelopes, err := d.QueryAll(ctx, id)
			if err != nil {
				return nil, err
			}
			return events.AggregateEnvelopes(envelopes)
		}
		return nil, err
	}
	params := dynamodb.QueryInput{
		TableName:              aws.String(d.Table),
		KeyConditionExpression: aws.String("Id = :uuid AND Version >= :version"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid":    &types.AttributeValueMemberS{Value: id},
			":version": &types.AttributeValueMemberN{Value: strconv.Itoa(snapshot.LatestVersion)},
		},
	}
	ml, err := d.query(ctx, &params)
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

func (d *DynamoDBEventStore) QueryLatestVersion(ctx context.Context, id string) (int, error) {
	params := dynamodb.QueryInput{
		TableName:              aws.String(d.Table),
		KeyConditionExpression: aws.String("Id = :uuid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id},
		},
		Limit:            aws.Int32(1),
		ScanIndexForward: aws.Bool(false),
	}
	var e []events.Snapshot
	mapList, err := d.query(ctx, &params)
	if err != nil {
		return 0, err
	}
	err = attributevalue.UnmarshalListOfMaps(mapList, &e)

	if err != nil {
		return 0, err
	}
	return e[0].Version, nil
}

// QueryAll takes a context and id and returns a slice of Events and an error
func (d *DynamoDBEventStore) QueryAll(ctx context.Context, id string) ([]events.Envelope, error) {
	params := dynamodb.QueryInput{
		TableName:              aws.String(d.Table),
		KeyConditionExpression: aws.String("Id = :uuid"),
		FilterExpression:       aws.String("Note <> :note"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id},
			":note": &types.AttributeValueMemberS{Value: SnapshotValue},
		},
	}
	maplist, err := d.query(ctx, &params)
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

func (d *DynamoDBEventStore) getLatestSnapshot(ctx context.Context, id string) (*events.Snapshot, error) {
	params := dynamodb.QueryInput{
		TableName:              aws.String(d.Table),
		KeyConditionExpression: aws.String("Id = :uuid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uuid": &types.AttributeValueMemberS{Value: id + SnapshotValue},
		},
		Limit:            aws.Int32(1),
		ScanIndexForward: aws.Bool(false),
	}
	var snapshots []events.Snapshot
	mapList, err := d.query(ctx, &params)
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
func (d *DynamoDBEventStore) query(ctx context.Context, params *dynamodb.QueryInput) (AttributeValueMapList, error) {
	var maps AttributeValueMapList
	// Query paginator provides pagination for queries until there are no more pages for DynamoDB to go through
	// See: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.Pagination.htm
	p := dynamodb.NewQueryPaginator(d.DB, params)
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

func (d *DynamoDBEventStore) append(ctx context.Context, valueMap AttributeValueMap) error {
	//This condition makes sure the sort key Version does not already exist
	input := &dynamodb.PutItemInput{
		TableName:           &d.Table,
		Item:                valueMap,
		ConditionExpression: aws.String("attribute_not_exists(Version)"),
	}
	_, err := d.DB.PutItem(ctx, input)
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
