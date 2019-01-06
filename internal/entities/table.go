package entities

import (
	"context"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/pkg/errors"
)

//A TableRequest is request which can be made to the table set
type TableRequest interface {
	Process(*TableSet) error
}

//TableSet is a set containing all tables
//The TableSet has it's own goroutine which can be used to query data from the tables
type TableSet struct {
	nodeTable NodeTable
	Channel   chan TableRequest
}

//Run starts a goroutine which is used to interact with the tables
func (set *TableSet) Run(ctx context.Context) error {
	for {
		select {
		case req := <-set.Channel:
			err := req.Process(set)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

//EventUpdate creates a new update table request and queues it
func (set *TableSet) EventUpdate(event Event) {
	set.Channel <- &UpdateTableRequest{
		Event: event,
	}
}

type UpdateTableRequest struct {
	Event Event
}

func (req *UpdateTableRequest) Process(tables *TableSet) error {

	switch event := req.Event.(type) {
	case *GenericEvent:
		eventType := event.UpdateType
		switch tableEntity := event.GetTableEntity().(type) {
		case *abusemesh.TableEvent_Node:
			err := tables.nodeTable.handleTableEvent(eventType, tableEntity)
			if err != nil {
				return err
			}
		default:
			return errors.Errorf("Unknown event type '%T'", event)
		}
	default:
		return errors.Errorf("Unknown event type '%T'", event)
	}

	return nil
}
