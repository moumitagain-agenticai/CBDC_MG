package transaction

import (
    "fmt"
    "sync"
)

// StateHandler defines a handler for a specific state
type StateHandler func(ctx context.Context, tx *Transaction) (*Transaction, error)

// StateMachine manages transaction state transitions
type StateMachine struct {
    transitions map[TransactionState]map[TransactionState]bool
    handlers    map[TransactionState]StateHandler
    mu          sync.RWMutex
}

// NewStateMachine creates a new state machine
func NewStateMachine() *StateMachine {
    sm := &StateMachine{
        transitions: make(map[TransactionState]map[TransactionState]bool),
        handlers:    make(map[TransactionState]StateHandler),
    }

    // Define valid transitions
    sm.addTransition(StateInitiated, StateComplianceCheck)
    sm.addTransition(StateComplianceCheck, StateFXConversion)
    sm.addTransition(StateComplianceCheck, StateFailed)
    sm.addTransition(StateFXConversion, StateLockFunds)
    sm.addTransition(StateFXConversion, StateFailed)
    sm.addTransition(StateLockFunds, StateSettlement)
    sm.addTransition(StateLockFunds, StateFailed)
    sm.addTransition(StateSettlement, StateReleaseLock)
    sm.addTransition(StateSettlement, StateFailed)
    sm.addTransition(StateReleaseLock, StateConfirmCompletion)
    sm.addTransition(StateReleaseLock, StateFailed)
    sm.addTransition(StateConfirmCompletion, StateCompleted)
    sm.addTransition(StateConfirmCompletion, StateFailed)

    return sm
}

// addTransition adds a valid transition
func (sm *StateMachine) addTransition(from, to TransactionState) {
    if sm.transitions[from] == nil {
        sm.transitions[from] = make(map[TransactionState]bool)
    }
    sm.transitions[from][to] = true
}

// RegisterHandler registers a handler for a specific state
func (sm *StateMachine) RegisterHandler(state TransactionState, handler StateHandler) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    sm.handlers[state] = handler
}

// GetHandler gets the handler for a specific state
func (sm *StateMachine) GetHandler(state TransactionState) (StateHandler, bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    handler, exists := sm.handlers[state]
    return handler, exists
}

// Transition transitions the transaction to a new state
func (sm *StateMachine) Transition(ctx context.Context, tx *Transaction, newState TransactionState) (*Transaction, error) {
    sm.mu.RLock()
    allowed := sm.isValidTransition(tx.State, newState)
    sm.mu.RUnlock()

    if !allowed {
        return nil, NewTransitionError(tx.State, newState)
    }

    // Update the transaction state
    tx.State = newState
    tx.Status = stateToStatus(newState)
    tx.UpdatedAt = time.Now()

    return tx, nil
}

// isValidTransition checks if a transition is valid
func (sm *StateMachine) isValidTransition(from, to TransactionState) bool {
    if targets, exists := sm.transitions[from]; exists {
        return targets[to]
    }
    return false
}

// GetNextState determines the next state based on the current state
func (sm *StateMachine) GetNextState(current TransactionState) (TransactionState, error) {
    switch current {
    case StateInitiated:
        return StateComplianceCheck, nil
    case StateComplianceCheck:
        return StateFXConversion, nil
    case StateFXConversion:
        return StateLockFunds, nil
    case StateLockFunds:
        return StateSettlement, nil
    case StateSettlement:
        return StateReleaseLock, nil
    case StateReleaseLock:
        return StateConfirmCompletion, nil
    case StateConfirmCompletion:
        return StateCompleted, nil
    case StateCompleted, StateFailed:
        return current, nil
    default:
        return current, fmt.Errorf("unknown state: %s", current)
    }
}

// IsTerminalState checks if a state is terminal
func (sm *StateMachine) IsTerminalState(state TransactionState) bool {
    return state == StateCompleted || state == StateFailed
}

// stateToStatus maps a state to a status
func stateToStatus(state TransactionState) TransactionStatus {
    switch state {
    case StateInitiated:
        return StatusPending
    case StateComplianceCheck:
        return StatusComplianceCheck
    case StateFXConversion:
        return StatusFXProcessing
    case StateLockFunds:
        return StatusLocking
    case StateSettlement:
        return StatusSettling
    case StateReleaseLock:
        return StatusReleasingLock
    case StateConfirmCompletion:
        return StatusPending
    case StateCompleted:
        return StatusCompleted
    case StateFailed:
        return StatusFailed
    default:
        return StatusPending
    }
}