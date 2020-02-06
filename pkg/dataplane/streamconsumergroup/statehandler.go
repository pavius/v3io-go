package streamconsumergroup

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/v3io/v3io-go/pkg/common"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type streamConsumerGroupStateHandler struct {
	logger                     logger.Logger
	streamConsumerGroup        *streamConsumerGroup
	lastState                  *State
	stopStateRefreshingChannel chan bool
}

func newStreamConsumerGroupStateHandler(streamConsumerGroup *streamConsumerGroup) (StateHandler, error) {
	return &streamConsumerGroupStateHandler{
		logger:                     streamConsumerGroup.logger.GetChild("stateHandler"),
		streamConsumerGroup:        streamConsumerGroup,
		stopStateRefreshingChannel: make(chan bool),
	}, nil
}

func (sh *streamConsumerGroupStateHandler) Start() error {
	state, err := sh.refreshState()
	if err != nil {
		return errors.Wrap(err, "Failed first refreshing state")
	}
	sh.lastState = state

	go sh.refreshStatePeriodically(sh.stopStateRefreshingChannel, sh.streamConsumerGroup.config.State.Heartbeat.Interval)

	return nil
}

func (sh *streamConsumerGroupStateHandler) Stop() error {
	sh.stopStateRefreshingChannel <- true
	return nil
}

func (sh *streamConsumerGroupStateHandler) refreshStatePeriodically(stopStateRefreshingChannel chan bool,
	heartbeatInterval time.Duration) {
	ticker := time.NewTicker(heartbeatInterval)

	for {
		select {
		case <-stopStateRefreshingChannel:
			ticker.Stop()
			return
		case <-ticker.C:
			state, err := sh.refreshState()
			if err != nil {
				sh.logger.WarnWith("Failed refreshing state", "err", errors.GetErrorStackString(err, 10))
				continue
			}
			sh.lastState = state
		}
	}
}

func (sh *streamConsumerGroupStateHandler) refreshState() (*State, error) {
	newState, err := sh.modifyState(func(state *State) (*State, error) {
		now := time.Now()
		if state == nil {
			state = &State{
				SchemasVersion: "0.0.1",
				Sessions:       make([]SessionState, 0),
			}
		}

		// remove stale sessions
		validSessions := make([]SessionState, 0)
		for index, sessionState := range state.Sessions {

			sessionTimeout := sh.streamConsumerGroup.config.Session.Timeout
			if !now.After(sessionState.LastHeartbeat.Add(sessionTimeout)) {
				validSessions = append(validSessions, state.Sessions[index])
			}
		}
		state.Sessions = validSessions

		// create or update sessions in the state
		for _, member := range sh.streamConsumerGroup.members {
			var sessionState *SessionState
			for index, session := range state.Sessions {
				if session.MemberID == member.ID {
					session = state.Sessions[index]
					break
				}
			}
			if sessionState == nil {
				if state.Sessions == nil {
					state.Sessions = make([]SessionState, 0)
				}
				shards, err := sh.resolveShardsToAssign(state)
				if err != nil {
					return nil, errors.Wrap(err, "Failed resolving shards for session")
				}
				state.Sessions = append(state.Sessions, SessionState{
					MemberID:      member.ID,
					LastHeartbeat: &now,
					Shards:        shards,
				})
			}
			sessionState.LastHeartbeat = &now
		}

		return state, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed modifying state")
	}

	return newState, nil
}

type stateModifier func(*State) (*State, error)

func (sh *streamConsumerGroupStateHandler) resolveShardsToAssign(state *State) ([]int, error) {
	numberOfShards, err := sh.resolveNumberOfShards()
	if err != nil {
		return nil, errors.Wrap(err, "Failed resolving number of shards")
	}

	maxNumberOfShardsPerSession, err := sh.resolveMaxNumberOfShardsPerSession(numberOfShards, sh.streamConsumerGroup.maxWorkers)
	if err != nil {
		return nil, errors.Wrap(err, "Failed resolving max number of shards per session")
	}

	shardIDs := common.MakeRange(0, numberOfShards-1)
	shardsToAssign := make([]int, 0)
	for _, shardID := range shardIDs {
		found := false
		for _, session := range state.Sessions {
			if common.IntSliceContainsInt(session.Shards, shardID) {
				found = true
				break
			}
		}
		if found {

			// sanity - it gets inside when there was unassigned shard but an assigned shard found filling the shards
			// list for the session or reaching end of shards list
			if len(shardsToAssign) > 0 {
				return nil, errors.New("Shards assignment out of order")
			}
			continue
		}
		shardsToAssign = append(shardsToAssign, shardID)
		if len(shardsToAssign) == maxNumberOfShardsPerSession {
			return shardsToAssign, nil
		}
	}

	// all shards are assigned
	if len(shardsToAssign) == 0 {
		// TODO: decide what to do
	}

	return shardsToAssign, nil
}

func (sh *streamConsumerGroupStateHandler) resolveMaxNumberOfShardsPerSession(numberOfShards int, maxWorkers int) (int, error) {
	if numberOfShards%maxWorkers != 0 {
		return numberOfShards/maxWorkers + 1, nil
	}
	return numberOfShards / maxWorkers, nil
}

func (sh *streamConsumerGroupStateHandler) resolveNumberOfShards() (int, error) {
	// TODO: implement this using list dir on stream dir and count files
	return 8, nil
}

func (sh *streamConsumerGroupStateHandler) modifyState(modifier stateModifier) (*State, error) {
	var modifiedState *State

	stateFilePath, err := sh.getStateFilePath()
	if err != nil {
		return nil, errors.Wrap(err, "Failed getting state file path")
	}

	backoff := sh.streamConsumerGroup.config.State.ModifyRetry.Backoff
	attempts := sh.streamConsumerGroup.config.State.ModifyRetry.Attempts
	ctx := context.TODO()

	err = common.RetryFunc(ctx, sh.logger, attempts, nil, &backoff, func(_ int) (bool, error) {
		var statePtr *State

		response, err := sh.streamConsumerGroup.container.GetItemSync(&v3io.GetItemInput{
			DataPlaneInput: sh.streamConsumerGroup.dataPlaneInput,
			Path:           stateFilePath,
			AttributeNames: []string{"__mtime", "state"},
		})
		if err != nil {
			errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
			if !errHasStatusCode {
				return true, errors.Wrap(err, "Got error without status code")
			}
			if errWithStatusCode.StatusCode() != 404 {
				return true, errors.Wrap(err, "Failed getting state item")
			}
		} else {
			getItemOutput := response.Output.(*v3io.GetItemOutput)

			stateContentsInterface, foundStateAttribute := getItemOutput.Item["state"]
			if !foundStateAttribute {
				return true, errors.New("Failed getting state attribute")
			}
			stateContents, ok := stateContentsInterface.(string)
			if !ok {
				return true, errors.New("Unknown type for state attribute")
			}

			var state State

			err = json.Unmarshal([]byte(stateContents), state)
			if err != nil {
				return true, errors.Wrap(err, "Failed unmarshaling state contents")
			}

			statePtr = &state
		}

		modifiedState, err := modifier(statePtr)
		if err != nil {
			return true, errors.Wrap(err, "Failed modifying state")
		}

		//modifiedStateContents, err := json.Marshal(modifiedState)
		_, err = json.Marshal(modifiedState)

		// TODO: create or update state (need to get mtime attribute from the item as well)

		return false, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed modifying state, attempts exhausted")
	}
	return modifiedState, nil
}

func (sh *streamConsumerGroupStateHandler) getStateFilePath() (string, error) {
	return path.Join(sh.streamConsumerGroup.streamPath, fmt.Sprintf("%s-state.json", sh.streamConsumerGroup.ID)), nil
}

func (sh *streamConsumerGroupStateHandler) GetState() (*State, error) {
	return sh.lastState, nil
}

func (sh *streamConsumerGroupStateHandler) GetMemberState(memberID string) (*SessionState, error) {
	state, err := sh.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Failed getting state")
	}
	for index, sessionState := range state.Sessions {
		if sessionState.MemberID == memberID {
			return &state.Sessions[index], nil
		}
	}
	return nil, errors.Errorf("Member state not found: %s", memberID)
}
