package main

type Destination interface {
	GetDestination(opts interface{}) string
}

type TopicDestination struct {
	Value string
}

func (t *TopicDestination) GetDestination(opts interface{}) string {
	return t.Value
}
