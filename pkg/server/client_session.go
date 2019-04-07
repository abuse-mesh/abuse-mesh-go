package server

import (
	"sync"

	"github.com/abuse-mesh/abuse-mesh-go/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type clientState int

const (
	//clientStateIdle is the initial state, we have registered that a node wants to be our client but we have no active stream
	//This state can mean that a connection is administratively down or that we have lost connection to the client
	clientStateIdle clientState = iota

	//clientStateEstablished indicates that we have a successful connection with a client
	//Moving from the idle state to the established state means that a full sync has to occur
	//Moving from the interrupted state to the established state means that only a partial state has to happen
	clientStateEstablished

	//clientStateInterrupted indicates that the connection was interrupted
	//We will wait for the client to reconnect, if the timeout is reached or the event buffer is full we have to move the state to idle
	clientStateInterrupted
)

//A session we are having with a client which is owned by the server
type clientSession struct {
	id     uuid.UUID
	state  clientState
	client *entities.Node
	// eventBuffer EventBuffer // The event buffer holds the last x events which can be resent to the client of it requests them
	eventCounter uint64
}

//clientSessionStorage stores client sessions
type clientSessionStorage struct {
	//All client sessions indexed on Node.UUID of the client
	sessions map[uuid.UUID]*clientSession

	//The mutex lock which prevents race conditions in sessions
	lock sync.RWMutex
}

//GetSession returns the session for client with the given uuid, if no session exist nil will be returned
func (storage *clientSessionStorage) GetSession(clientID uuid.UUID) *clientSession {
	storage.lock.RLock()
	defer storage.lock.RUnlock()
	return storage.sessions[clientID]
}

func (storage *clientSessionStorage) RemoveSession(session *clientSession) error {
	if session.client == nil {
		return errors.New("Client session can't be nil")
	}

	storage.lock.Lock()
	defer storage.lock.Unlock()

	delete(storage.sessions, session.client.UUID)

	return nil
}

func (storage *clientSessionStorage) AddSession(session *clientSession) error {
	if session.client == nil {
		return errors.New("Client session can't be nil")
	}

	if storage.GetSession(session.client.UUID) != nil {
		return errors.Errorf("Client '%s' already has a session", session.client.UUID)
	}

	storage.lock.Lock()
	defer storage.lock.Unlock()

	storage.sessions[session.client.UUID] = session

	return nil
}
