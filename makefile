generate-protos:
	./scripts/proto_gen.sh

unit-tests:
	go test ./... -v -short

tests:
	go test ./... -v