package model

type Event interface {
	GetId() int
	GetKey() string
}
