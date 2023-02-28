package events

import "github.com/cpustejovsky/event-store/protos/hitpoints"

type Aggregator interface {
	Aggregate([][]byte) ([]byte, error)
}

var HitPointsName string = string(hitpoints.File_protos_hitpoints_hitpoints_proto.FullName().Name())

type EventsMap map[string]Aggregator

func NewEventMap() EventsMap {
	return EventsMap{
		HitPointsName: &HitPoints{},
	}
}
