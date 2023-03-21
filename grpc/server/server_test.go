package server_test

import (
	"context"
	"github.com/cpustejovsky/event-store/events"
	"github.com/cpustejovsky/event-store/grpc/server"
	pb "github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

var EventStoreTable = "event-store"
var id = "ab9261dc-685c-4cec-b6f0-f6fb886cdf7cs"

type StubEventStore struct {
	Appended             bool
	QueriedLatestVersion bool
}

func (s *StubEventStore) Append(context.Context, *events.Envelope) error {
	s.Appended = true
	return nil
}
func (s *StubEventStore) Snapshot(context.Context, *events.Snapshot) error {
	return nil
}
func (s *StubEventStore) Project(context.Context, string) (*events.Envelope, error) {
	return nil, nil
}
func (s *StubEventStore) QueryLatestVersion(context.Context, string) (int, error) {
	s.QueriedLatestVersion = true
	return 0, nil

}
func (s *StubEventStore) QueryAll(context.Context, string) ([]events.Envelope, error) {
	return nil, nil
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestNewServer(t *testing.T) {
	ctx := context.TODO()
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	//Create Envelope Store
	es := &StubEventStore{}
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
		Id:                 id,
		CharacterName:      "cpustejovsky",
		CharacterHitPoints: 8,
		Note:               "initialization",
	}
	_, err = c.RecordHitPoints(ctx, &pchp)
	assert.Nil(t, err)
	assert.True(t, es.Appended)
	assert.True(t, es.QueriedLatestVersion)
}
