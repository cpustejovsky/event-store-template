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
	"log"
	"os"
	"testing"
)

var EventStoreTable = "event-store"
var hp int32
var client *dynamodb.Client
var ctx = context.Background()
var id = "1aa75e80-51e4-48d9-a5b7-2f5e49e78e86"
var name = "cpustejovsky"

func init() {
	//Create Dynamodb Client
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
	if err != nil {
		log.Fatal(err)
	}
	client = dynamodb.NewFromConfig(cfg)
}

func TestEventStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test")
	}

	//Create Event Store
	es := store.DynamoDB(client, EventStoreTable)
	require.NotNil(t, es)

	hitPointEvents := []hitpoints.PlayerCharacterHitPoints{{
		Id:                 id,
		CharacterName:      name,
		CharacterHitPoints: 8,
		Note:               "Init",
	}, {
		Id:                 id,
		CharacterName:      name,
		CharacterHitPoints: -2,
		Note:               "Slashing damage from goblin",
	}, {
		Id:                 id,
		CharacterName:      name,
		CharacterHitPoints: -3,
		Note:               "bludgeoning damage from bugbear",
	}}
	var envelopes []events.Envelope
	for i, e := range hitPointEvents {
		hp += e.GetCharacterHitPoints()
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

	t.Run("Append Items to Envelope Store", func(t *testing.T) {
		for _, event := range envelopes {
			err := es.Append(ctx, &event)
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("Attempt to append existing version to event store and fail", func(t *testing.T) {
		e := events.Envelope{
			Id:        id,
			Version:   0,
			EventName: events.HitPointsName,
			Event:     []byte{},
		}
		err := es.Append(ctx, &e)
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
		require.Nil(t, err)
		snapshot := &events.Snapshot{
			Id:            agg.Id,
			LatestVersion: agg.Version,
			Event:         agg.Event,
			EventName:     agg.EventName,
		}
		err = es.Snapshot(ctx, snapshot)
		assert.Nil(t, err)
	})

	t.Run("QueryAll should not return the snapshot", func(t *testing.T) {
		queriedEvents, err := es.QueryAll(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, len(envelopes), len(queriedEvents))
	})

	t.Cleanup(func() {
		p := dynamodb.NewScanPaginator(client, &dynamodb.ScanInput{
			TableName: aws.String(EventStoreTable),
		})
		for p.HasMorePages() {
			out, err := p.NextPage(ctx)
			if err != nil {
				t.Fatal(err)
			}

			for _, item := range out.Items {
				_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
					TableName: aws.String(EventStoreTable),
					Key: map[string]types.AttributeValue{
						"Id":      item["Id"],
						"Version": item["Version"],
					},
				})
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}
