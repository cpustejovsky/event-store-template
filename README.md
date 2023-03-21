# event-store-template

This is an event store template.

## Layout

The Event Store interface has the following methods:
```go
type EventStore interface {
	Append(context.Context, *events.Envelope) error
	Snapshot(context.Context, *events.Snapshot) error
	Project(context.Context, string) (*events.Envelope, error)
	QueryLatestVersion(context.Context, string) (int, error)
	QueryAll(context.Context, string) ([]events.Envelope, error)
}
```

This is the `Envelope` type:
```go
// Envelope contains necessary information to store event in the event store
type Envelope struct {
	Id        string
	Version   int
	Event     []byte
	EventName string
}
```

This is the `Snapshot` type:
```go
// Snapshot contains aggregated event information along with last version
type Snapshot struct {
	Id            string
	Version       int
	LatestVersion int
	Event         []byte
	EventName     string
}
```

**NOTE: `EventName` is a specific string provided by the protobuf generated file.**
For example, the `EventName` for `./protos/hitpoints/hitpoints.proto` is:

```go
package main

import (
  hitpointspb "github.com/cpustejovsky/event-store/protos/hitpoints"
)

var HitPointsName string = string(hitpointspb.File_protos_hitpoints_hitpoints_proto.FullName().Name())
```

## Use

It currently has a specific DynamoDB implementation that takes a DynamoDB client:
```go
cfg, err := config.LoadDefaultConfig(context.TODO)
client := dynamodb.NewFromConfig(cfg)
es := store.DynamoDB(client, "event-store-table-name")
```

## Testing
* To run unit tests, run `make unit-tests`
* To run all tests, run `make tests`
  * Make sure to have the follow environment variables set to run integration tests
    * `AWS_REGION`
    * `AWS_ACCESS_ID`
    * `AWS_SECRET_KEY`

## TODOS
* Add ability to configure `EventsMap` and pass in a custom events map to the event store 