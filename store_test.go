package store_test

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	store "github.com/cpustejovsky/event-store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
	"time"
)

var EventStoreTable = "event-store"

func TestEventStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test")
	}

	//Create Dynamodb Client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				},
			),
		),
		config.WithCredentialsProvider(
			credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     os.Getenv("AWS_ACCESS_ID"),
					SecretAccessKey: os.Getenv("AWS_SECRET_KEY"),
				},
			},
		),
	)
	require.Nil(t, err)

	//Create Event Store
	client := dynamodb.NewFromConfig(cfg)
	es := store.New(client, EventStoreTable)
	require.NotNil(t, es)
	id := uuid.NewString()
	name := "cpustejovsky"
	events := []store.Event{
		{
			Id:                      id,
			Version:                 0,
			CharacterName:           name,
			CharacterHitPointChange: 8,
			Note:                    "Init",
		},
		{
			Id:                      id,
			Version:                 1,
			CharacterName:           name,
			CharacterHitPointChange: -2,
			Note:                    "Slashing damage from goblin",
		},
		{
			Id:                      id,
			Version:                 2,
			CharacterName:           name,
			CharacterHitPointChange: -3,
			Note:                    "bludgeoning damage from bugbear",
		},
	}
	t.Run("Append Items to Event Store", func(t *testing.T) {
		for _, event := range events {
			err := es.Append(context.Background(), &event)
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("Attempt to append existing version to event store and fail", func(t *testing.T) {
		e := store.Event{
			Id:      id,
			Version: 0,
		}
		err := es.Append(context.Background(), &e)
		assert.NotNil(t, err)
		checkErr := &store.EventAlreadyExistsError{}
		assert.True(t, errors.As(err, &checkErr))
	})

	t.Run("QueryAll Items from Event Store", func(t *testing.T) {
		queriedEvents, err := es.QueryAll(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, len(events), len(queriedEvents))
		for _, event := range events {
			assert.Contains(t, queriedEvents, event)
		}
	})

	t.Run("QueryAll returns specific error if no Event is found", func(t *testing.T) {
		_, err := es.QueryAll(ctx, uuid.NewString())
		assert.NotNil(t, err)
		checkErr := &store.NoEventFoundError{}
		assert.True(t, errors.As(err, &checkErr))
	})

	t.Run("QuerySinceVersion from Event Store", func(t *testing.T) {
		queriedEvents, err := es.QuerySinceVersion(ctx, id, 1)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(queriedEvents))

	})

	t.Run("Attempt QuerySinceVersion from Event Store with version higher than in event store", func(t *testing.T) {
		events, err := es.QuerySinceVersion(ctx, id, 42)
		assert.NotNil(t, err)
		assert.Nil(t, events)
	})

	t.Run("Replay Events from Event Store", func(t *testing.T) {
		aggEvent, err := es.Replay(ctx, id)
		require.Nil(t, err)
		assert.Nil(t, err)
		hp := events[0].CharacterHitPointChange + events[1].CharacterHitPointChange + events[2].CharacterHitPointChange
		assert.Equal(t, hp, aggEvent.CharacterHitPoints)
		assert.Equal(t, name, aggEvent.CharacterName)
	})

	t.Run("QueryFromLastSnapshot returns all events if there is no snapshot", func(t *testing.T) {
		queriedEvents, err := es.QueryFromLastSnapshot(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, len(events), len(queriedEvents))
		for _, event := range events {
			assert.Contains(t, queriedEvents, event)
		}
	})

	t.Cleanup(func() {
		for i := 0; i < len(events); i++ {
			params := dynamodb.DeleteItemInput{
				TableName: &EventStoreTable,
				Key: map[string]types.AttributeValue{
					"Id":      &types.AttributeValueMemberS{Value: id},
					"Version": &types.AttributeValueMemberN{Value: strconv.Itoa(i)},
				},
			}
			_, err := client.DeleteItem(context.Background(), &params)
			if err != nil {
				t.Log("Error deleting items for cleanup:\t", err)
			}
		}
	})
}
