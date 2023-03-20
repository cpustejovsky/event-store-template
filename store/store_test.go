package store_test

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cpustejovsky/event-store/events"
	"github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/cpustejovsky/event-store/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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

	//Create Envelope Store
	client := dynamodb.NewFromConfig(cfg)
	es := store.New(client, EventStoreTable)
	require.NotNil(t, es)
	id := uuid.NewString()
	name := "cpustejovsky"
	hitPointEvents := []hitpoints.PlayerCharacterHitPoints{{
		CharacterName:      name,
		CharacterHitPoints: 8,
		Note:               "Init",
	}, {
		CharacterName:      name,
		CharacterHitPoints: -2,
		Note:               "Slashing damage from goblin",
	}, {
		CharacterName:      name,
		CharacterHitPoints: -3,
		Note:               "bludgeoning damage from bugbear",
	}}
	var envelopes []events.Envelope
	for i, e := range hitPointEvents {
		bin, err := proto.Marshal(&e)
		if err != nil {
			t.Fatal(err)
		}
		envelopes = append(envelopes, events.Envelope{
			Id:        id,
			Version:   i,
			Event:     bin,
			EventName: events.HitPointsName,
		})
	}

	hp := hitPointEvents[0].CharacterHitPoints + hitPointEvents[1].CharacterHitPoints + hitPointEvents[2].CharacterHitPoints
	t.Run("Append Items to Envelope Store", func(t *testing.T) {
		for _, event := range envelopes {
			err := es.Append(context.Background(), &event)
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("Attempt to append existing version to event store and fail", func(t *testing.T) {
		e := events.Envelope{
			Id:        id,
			Version:   0,
			EventName: "test",
			Event:     []byte{},
			Note:      "",
		}
		err := es.Append(context.Background(), &e)
		assert.NotNil(t, err)
		checkErr := &store.EventAlreadyExistsError{}
		assert.True(t, errors.As(err, &checkErr))
	})

	t.Run("QueryAll Items from Envelope Store", func(t *testing.T) {
		queriedEvents, err := es.QueryAll(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, len(envelopes), len(queriedEvents))
		for _, event := range envelopes {
			assert.Contains(t, queriedEvents, event)
		}
	})

	t.Run("QueryAll returns specific error if no Envelope is found", func(t *testing.T) {
		_, err := es.QueryAll(ctx, uuid.NewString())
		assert.NotNil(t, err)
		checkErr := &store.NoEventFoundError{}
		assert.True(t, errors.As(err, &checkErr))
	})

	t.Run("Project Events from Envelope Store since Snapshot", func(t *testing.T) {
		agg, err := es.Project(ctx, id)
		require.Nil(t, err)
		hpEvent := hitpoints.PlayerCharacterHitPoints{}
		err = proto.Unmarshal(agg.Event, &hpEvent)
		require.Nil(t, err)
		assert.Equal(t, hp, hpEvent.CharacterHitPoints)
		assert.Equal(t, name, hpEvent.CharacterName)
	})

	t.Run("Snapshot should return no error", func(t *testing.T) {
		agg, err := es.Project(ctx, id)
		assert.Nil(t, err)
		err = es.Snapshot(ctx, agg)
		assert.Nil(t, err)
	})

	t.Run("QueryAll should not return the snapshot", func(t *testing.T) {
		queriedEvents, err := es.QueryAll(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, len(envelopes), len(queriedEvents))
		for _, e := range envelopes {
			assert.NotEqual(t, store.SnapshotValue, e.Note)
		}
	})

	t.Cleanup(func() {
		for i := 0; i < len(envelopes); i++ {
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
