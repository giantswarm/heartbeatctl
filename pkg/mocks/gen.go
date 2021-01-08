package mocks

//go:generate go run github.com/golang/mock/mockgen -source=../client/ports.go -mock_names=Port=MockedClient -package=mocks -destination=./client.go
