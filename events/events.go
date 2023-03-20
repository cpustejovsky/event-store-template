package events

import (
	hitpointspb "github.com/cpustejovsky/event-store/protos/hitpoints"
	levelspb "github.com/cpustejovsky/event-store/protos/levels"
)

type Aggregator interface {
	Aggregate([][]byte) ([]byte, error)
}

var HitPointsName string = string(hitpointspb.File_protos_hitpoints_hitpoints_proto.FullName().Name())
var LevelsName string = string(levelspb.File_protos_levels_levels_proto.FullName().Name())

type EventsMap map[string]Aggregator

func NewEventMap() EventsMap {
	return EventsMap{
		HitPointsName: &HitPoints{},
		LevelsName:    &Levels{},
	}
}

// Envelope contains necessary information to store event in the event store
type Envelope struct {
	Id        string
	Version   int
	Event     []byte
	EventName string
	Note      string
}

func AggregateEnvelopes(envelopes []Envelope) (*Envelope, error) {
	var e Envelope
	//Loop through envelopes to get
	var events [][]byte
	for i, envelope := range envelopes {
		if i == len(envelopes)-1 {
			e.Id = envelope.Id
			e.Version = envelope.Version + 1
			e.EventName = envelope.EventName
		}
		events = append(events, envelope.Event)
	}
	agg, err := reconstituteEvents(events, e.EventName)
	if err != nil {
		return nil, err
	}
	e.Event = agg
	return &e, nil
}

func reconstituteEvents(es [][]byte, name string) ([]byte, error) {
	//map name to wrapper of protobuf type that has a reconstitute method
	m := NewEventMap()
	return m[name].Aggregate(es)
}
