package server_test

import (
	"context"
	"github.com/cpustejovsky/event-store/grpc/server"
	pb "github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestNewServer(t *testing.T) {
	ctx := context.TODO()
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	svr := server.New()
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
	t.Log(err)
}
