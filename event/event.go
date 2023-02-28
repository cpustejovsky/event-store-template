package event

import "github.com/cpustejovsky/event-store/events"

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
	firstEnv := envelopes[0]
	e.Id = firstEnv.Id
	e.Version = firstEnv.Version
	e.EventName = firstEnv.EventName
	//Loop through envelopes to get
	var events [][]byte
	for _, envelope := range envelopes {
		events = append(events, envelope.Event)
	}
	agg, err := reconstituteEvents(events, firstEnv.EventName)
	if err != nil {
		return nil, err
	}
	e.Event = agg
	return &e, nil
}

func reconstituteEvents(es [][]byte, name string) ([]byte, error) {
	//map name to wrapper of protobuf type that has a reconstitute method
	m := events.NewEventMap()
	return m[name].Aggregate(es)
}
