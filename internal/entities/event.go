package entities

import (
	"context"
	"sync"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/internal/utils/conv"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	//ErrEventEntityEmpty signals that the GRPC message had no TableEntity
	ErrEventEntityEmpty = errors.New("The event contains no entity data")
)

//Event is a change of data in a table
type Event interface {
	//Validate checks if the event message is valid
	//Returns true if the event is valid or false and a error containing the reason it is not
	Validate(*TableSet) (bool, error)

	//GetID returns the id of the event
	GetID() uuid.UUID
}

//GenericEvent is a wrapper for the protocol stub
type GenericEvent struct {
	abusemesh.TableEvent
}

//Validate checks if the event message is valid
//Returns true if the event is valid or false and a error containing the reason it is not
func (event *GenericEvent) Validate(tableSet *TableSet) (bool, error) {
	//The entity which is the subject of this event
	entity := event.GetTableEntity()

	_, err := conv.AuuidToGuuid(event.GetEventId())
	if err != nil {
		return false, errors.Wrap(err, "Event ID invalid")
	}

	switch e := entity.(type) {
	case *abusemesh.TableEvent_Node:
		//In case of a new node on the network we need to contact that node and confirm it's existence and claims
		return false, errors.New("Not Yet Implemented")

	case *abusemesh.TableEvent_Report:
		//In case of a report we need verify that the signature is correct
		return false, errors.New("Not Yet Implemented")

	case *abusemesh.TableEvent_ReportConfirmation:
		//In case of a report confirmation we need verify that the signature is correct
		return false, errors.New("Not Yet Implemented")

	case *abusemesh.TableEvent_DelistAcceptance:
		//In case of a delist acceptance we need verify that the signature is correct
		return false, errors.New("Not Yet Implemented")

	case *abusemesh.TableEvent_DelistRequests:
		//In case of a delist request we need verify that the signature is correct
		return false, errors.New("Not Yet Implemented")

	case *abusemesh.TableEvent_Neighbor:
		//In case of a neighbor we need verify that the signature is correct
		return false, errors.New("Not Yet Implemented")

	case nil:
		return false, ErrEventEntityEmpty

	default:
		logrus.Errorf("TableEvent.GetTableEntity() has unexpected type '%T'", e)
		return false, errors.New("Protocol error, check error log")
	}
}

//GetID returns the id of the event
func (event *GenericEvent) GetID() uuid.UUID {
	id, err := conv.AuuidToGuuid(event.GetEventId())
	if err != nil {
		return uuid.UUID{}
	}

	return id
}

//EventObserver specifies a struct which can receive event updates
type EventObserver interface {
	EventUpdate(Event)
}

//A EventStream holds all events known the node
type EventStream interface {
	//GetWriteChannel returns a channel which can be used by other components to write a new event to the stream
	//The EventStream is responsible for checking the validity and uniqueness of the event
	GetWriteChannel() chan<- Event

	//GetAllEvents returns all events currently in the event stream
	GetAllEvents() []Event

	//Attach can be used by other components to subscribe to updates of the event stream
	//The EventStream must only call the callback with validated and unique events.
	Attach(observerCallback EventObserver)

	//Detach removes a subscriber
	Detach(observerCallback EventObserver)

	//Run runs the goroutine which handles changes and requests to the EventStream
	Run(context.Context) error
}

//The inMemoryEventStream implements EventStream and stores the events in memory.
//In memory storage is fast but can also lead to excessive memory usage and long GC pauses
type inMemoryEventStream struct {
	//All events in the event stream
	events map[uuid.UUID]Event

	//A mutex lock for the events
	eventsLock sync.RWMutex

	//A map of observers interested in new events
	observers []EventObserver

	//A mutex lock for the observers
	observerLock sync.Mutex

	//A channel which can be used to write new attempt
	writeChan chan Event

	//A table set which can be used to query the node table
	tableSet *TableSet
}

//NewInMemoryEventStream creates a new in memory event stream
func NewInMemoryEventStream(tableSet *TableSet, writeChanBufferSize int) EventStream {
	return &inMemoryEventStream{
		events:       make(map[uuid.UUID]Event),
		eventsLock:   sync.RWMutex{},
		observerLock: sync.Mutex{},
		writeChan:    make(chan Event, writeChanBufferSize),
		tableSet:     tableSet,
	}
}

func (stream *inMemoryEventStream) GetWriteChannel() chan<- Event {
	return stream.writeChan
}

func (stream *inMemoryEventStream) Run(ctx context.Context) error {
	for {
		select {
		case event := <-stream.writeChan:
			valid, reason := event.Validate(stream.tableSet)
			if !valid {
				logrus.WithError(reason).Warn("Event refused because it is invalid")
				continue
			}

			eventID := event.GetID()

			//If the event doesn't already exist we add it to the stream and notify the observers
			if _, found := stream.events[eventID]; !found {
				stream.eventsLock.Lock()
				stream.events[eventID] = event
				stream.eventsLock.Unlock()

				stream.observerLock.Lock()
				for _, observer := range stream.observers {
					observer.EventUpdate(event)
				}
				stream.observerLock.Unlock()

			} else {
				logrus.WithField("event-id", eventID.String()).Info("Received duplicate event")
			}
		case <-ctx.Done():
			return nil
		}
	}
}

//Attach can be used by other components to subscribe to updates of the event stream
//Updates to the EventStream will be sent over the channel.
//The EventStream must only send validated and unique events.
func (stream *inMemoryEventStream) Attach(observer EventObserver) {
	//Lock the observers to avoid race conditions
	stream.observerLock.Lock()
	defer stream.observerLock.Unlock()

	stream.observers = append(stream.observers, observer)
}

//Detach removes a subscriber
func (stream *inMemoryEventStream) Detach(observer EventObserver) {
	//Lock the observers to avoid race conditions
	stream.observerLock.Lock()
	defer stream.observerLock.Unlock()

	//Find observer with the same value and delete it
	for index, curObserver := range stream.observers {
		if curObserver == observer {
			stream.observers = append(stream.observers[:index], stream.observers[index+1:]...)
			return
		}
	}
}

func (stream *inMemoryEventStream) GetAllEvents() []Event {
	//Lock the events to avoid race conditions
	stream.eventsLock.RLock()
	defer stream.eventsLock.RUnlock()

	events := make([]Event, 0, len(stream.events))
	for _, event := range stream.events {
		events = append(events, event)
	}

	return events
}
