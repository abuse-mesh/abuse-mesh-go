package server

import (
	"context"
	stdErrors "errors"
	"sync"
	"time"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/internal/entities"
	"github.com/abuse-mesh/abuse-mesh-go/internal/utils/conv"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/client"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	errSessionStopped = stdErrors.New("Context of session has signaled to stop")
)

type serverState int

const (
	//serverStateIdle is the initial state, we have registered that we want to be a client of a server but we have no active stream
	//This state can mean that a connection is administratively down or that we have lost connection to the server
	serverStateIdle serverState = iota
	//serverStateConnecting means that we are attempting to setup a event stream
	serverStateConnecting
	//serverStateEstablished means that we currently have a established connection
	serverStateEstablished
	//serverStateInterupted means that the connection was interupted
	serverStateInterupted
)

type serverSession struct {
	//The id of the session
	id uuid.UUID
	//Is it down because the admin disabled the session
	adminDown bool
	//The current state of the session
	state serverState
	//The server we are talking to
	server *entities.Node
	//The abuseMeshClient we are using for the connection
	abuseMeshClient *client.AbuseMeshClient
	//the eventStreamClient is the client object with which we can receive event from the server
	eventStreamClient abusemesh.AbuseMesh_TableEventStreamClient
	//eventStreamWriteChan is a channel provided by the local event stream on which we can publish new events
	eventStreamWriteChan chan<- entities.Event
	//The event counter
	eventCounter uint64
	//The context of the session
	context context.Context
}

//The control loop of the client
func (session *serverSession) Run() error {
	//The time when the next connection may be attempted
	nextConnAttempt := time.Now()
	//The amount of consecutive times a connection attempt has failed
	failedConnectionAttempts := 0
	//The amount of consecutive times a reconnect attempt has failed
	failedReconnectAttempts := 0
	//The cancel function of the event steam client
	var eventStreamClientCancel context.CancelFunc

	for {
		//If the session context is done
		if session.context.Err() != nil {
			//Stop and return
			return errSessionStopped
		}

		//Check the state
		switch session.state {
		case serverStateIdle: //if idle
			//If admin down
			if session.adminDown {
				//Sleep for a second
				time.Sleep(1 * time.Second)
				continue
			}

			//if the time of the next connection attempt is in the past we change state to 'connecting'
			if time.Until(nextConnAttempt) < 0 {
				//The next connection attempt can be after 30 seconds
				//TODO make backoff period configurable
				nextConnAttempt = time.Now().Add(30 * time.Second)

				//Change the state
				session.state = serverStateConnecting
				continue
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		case serverStateConnecting:
			negotiationResponse, err := session.abuseMeshClient.NegotiateNeighborship(&abusemesh.NegotiateNeighborshipRequest{})
			if err != nil {
				log.WithError(err).Error("Error while negotiating neighborship")

				failedConnectionAttempts++
				if failedConnectionAttempts >= 3 {
					session.state = serverStateIdle
				}

				time.Sleep(1 * time.Second)
				continue
			}

			session.id, err = conv.AuuidToGuuid(negotiationResponse.SessionId)
			if err != nil {
				log.WithError(err).Error("Protocol error: error while converting session uuid")

				failedConnectionAttempts++
				if failedConnectionAttempts >= 3 {
					session.state = serverStateIdle
					session.id = uuid.UUID{}
				}

				time.Sleep(1 * time.Second)
				continue
			}

			session.eventStreamClient, eventStreamClientCancel, err = session.abuseMeshClient.TableEventStream(&abusemesh.TableEventStreamRequest{
				Offset:    session.eventCounter,
				SessionId: &abusemesh.UUID{Uuid: session.id.String()},
			})

			if err != nil {
				eventStreamClientCancel()

				session.eventStreamClient = nil
				eventStreamClientCancel = nil

				log.WithError(err).Error("Error while opening event stream")

				failedConnectionAttempts++
				//If we have reached the connection attempt limit we return to idle
				//TODO make limit configurable
				if failedConnectionAttempts >= 3 {
					session.state = serverStateIdle
					failedConnectionAttempts = 0

					session.id = uuid.UUID{}
					session.eventCounter = 0

					//TODO make idle backoff time configurable
					nextConnAttempt = time.Now().Add(30 * time.Second)
				}

				time.Sleep(1 * time.Second)
				continue
			}

			//reset the failed connection counter
			failedConnectionAttempts = 0
			session.state = serverStateEstablished

		case serverStateEstablished:
			event, err := session.eventStreamClient.Recv()
			if err != nil {
				eventStreamClientCancel()

				session.eventStreamClient = nil
				eventStreamClientCancel = nil

				//NOTE should this be a error message? it may occur often
				//Maybe change it to a warning and error if we can't recover in the interupted state?
				log.WithError(err).Error("Event stream was closed")

				session.state = serverStateInterupted
				continue
			}

			//TODO add missing event check

			session.eventStreamWriteChan <- &entities.GenericEvent{TableEvent: *event}

			session.eventCounter++

		case serverStateInterupted:
			var err error
			session.eventStreamClient, eventStreamClientCancel, err = session.abuseMeshClient.TableEventStream(&abusemesh.TableEventStreamRequest{
				Offset:    session.eventCounter,
				SessionId: &abusemesh.UUID{Uuid: session.id.String()},
			})

			if err != nil {
				eventStreamClientCancel()

				session.eventStreamClient = nil
				eventStreamClientCancel = nil

				log.WithError(err).Error("Error while reconnecting to event stream")

				failedReconnectAttempts++
				//If we have reached the reconnect attempt limit we return to idle
				//TODO make limit configurable
				if failedReconnectAttempts >= 3 {
					session.state = serverStateIdle
					failedReconnectAttempts = 0

					session.id = uuid.UUID{}
					session.eventCounter = 0

					//TODO make idle backoff time configurable
					nextConnAttempt = time.Now().Add(30 * time.Second)
				}

				time.Sleep(1 * time.Second)
				continue
			}
		}
	}
}

type serverSessionStorage struct {
	sessions map[uuid.UUID]*serverSession
	lock     sync.RWMutex
}

//GetSession returns the session for client with the given uuid, if no session exist nil will be returned
func (storage *serverSessionStorage) GetSession(serverID uuid.UUID) *serverSession {
	storage.lock.RLock()
	defer storage.lock.RUnlock()
	return storage.sessions[serverID]
}

func (storage *serverSessionStorage) RemoveSession(session *serverSession) error {
	if session.abuseMeshClient == nil {
		return errors.New("Server session can't be nil")
	}

	storage.lock.Lock()
	defer storage.lock.Unlock()

	delete(storage.sessions, session.server.UUID)

	return nil
}

func (storage *serverSessionStorage) AddSession(session *serverSession) error {
	if session.abuseMeshClient == nil {
		return errors.New("Server session can't be nil")
	}

	if storage.GetSession(session.server.UUID) != nil {
		return errors.Errorf("Server '%s' already has a session", session.server.UUID)
	}

	storage.lock.Lock()
	defer storage.lock.Unlock()

	storage.sessions[session.server.UUID] = session

	return nil
}
