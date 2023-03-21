package server_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cpustejovsky/event-store/grpc/server"
	pb "github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/cpustejovsky/event-store/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

var EventStoreTable = "event-store"

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestNewServer(t *testing.T) {
	ctx := context.TODO()
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

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
	svr := server.New(es)
	pb.RegisterHitPointsRecorderServer(s, svr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(bufDialer))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHitPointsRecorderClient(conn)
	pchp := pb.PlayerCharacterHitPoints{
		CharacterName:      "cpustejovsky",
		CharacterHitPoints: 8,
		Note:               "initualization",
	}
	_, err = c.RecordHitPoints(ctx, &pchp)
	assert.Nil(t, err)
	t.Cleanup(func() {
		p := dynamodb.NewScanPaginator(client, &dynamodb.ScanInput{
			TableName: aws.String(EventStoreTable),
		})
		for p.HasMorePages() {
			out, err := p.NextPage(context.TODO())
			if err != nil {
				panic(err)
			}

			for _, item := range out.Items {
				_, err = client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
					TableName: aws.String(EventStoreTable),
					Key: map[string]types.AttributeValue{
						"Id":      item["Id"],
						"Version": item["Version"],
					},
				})
				if err != nil {
					log.Println(err)
					panic(err)
				}
			}
		}
	})
}
