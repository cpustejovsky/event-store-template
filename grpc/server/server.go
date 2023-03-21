package server

import (
	"context"
	pb "github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/golang/protobuf/ptypes/empty"
)

type Server struct {
	pb.UnimplementedHitPointsRecorderServer
}

func New() *Server {
	return &Server{}
}

func (s *Server) RecordHitPoints(ctx context.Context, hp *pb.PlayerCharacterHitPoints) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
