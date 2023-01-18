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
	"sync"
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
	numberOfEvents := 3

	t.Run("Append Items to Event Store", func(t *testing.T) {
		errChan := make(chan error, numberOfEvents)
		var wg sync.WaitGroup
		wg.Add(numberOfEvents)
		for i := 0; i < numberOfEvents; i++ {
			go func(v int) {
				defer wg.Done()
				e := store.Event{
					Id:      id,
					Version: v,
					Character: store.Character{
						Name:      "cpustejovsky",
						HitPoints: 42,
					},
					Note: "Test",
				}
				errChan <- es.Append(context.Background(), &e)
			}(i)
		}
		wg.Wait()
		close(errChan)
		for err := range errChan {
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
		events, err := es.QueryAll(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, numberOfEvents, len(events))
	})

	t.Run("Fail to query Items from Event Store", func(t *testing.T) {
		events, err := es.QueryAll(ctx, uuid.NewString())
		assert.NotNil(t, err)
		assert.Nil(t, events)
	})

	t.Run("QuerySinceVersion from Event Store", func(t *testing.T) {
		events, err := es.QuerySinceVersion(ctx, id, 1)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(events))

	})

	t.Run("Attempt QuerySinceVersion from Event Store with version higher than in event store", func(t *testing.T) {
		events, err := es.QuerySinceVersion(ctx, id, 42)
		assert.NotNil(t, err)
		assert.Nil(t, events)
	})

	t.Cleanup(func() {
		for i := 0; i < numberOfEvents; i++ {
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
