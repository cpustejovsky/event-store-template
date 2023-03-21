package server

import (
	"context"
	"errors"
	"github.com/cpustejovsky/event-store/events"
	pb "github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/cpustejovsky/event-store/store"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	Store *store.EventStore
	pb.UnimplementedHitPointsRecorderServer
}

func New(es *store.EventStore) *Server {
	return &Server{
		Store: es,
	}
}

func (s *Server) RecordHitPoints(ctx context.Context, hp *pb.PlayerCharacterHitPoints) (*empty.Empty, error) {
	bin, err := proto.Marshal(hp)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	v, err := s.Store.QueryLatestVersion(ctx, id)
	if err != nil && !errors.Is(err, &store.NoEventFoundError{}) {
		return nil, err
	}
	envelope := events.Envelope{
		Id:        id,
		Version:   v,
		Event:     bin,
		EventName: string(pb.File_protos_hitpoints_hitpoints_proto.FullName()),
	}
	err = s.Store.Append(ctx, &envelope)
	return &empty.Empty{}, err
}
