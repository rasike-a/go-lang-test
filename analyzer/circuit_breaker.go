package analyzer

import (
	"sync"
	"time"
)

// CircuitBreaker states
const (
	StateClosed = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	state           int
	failureCount    int
	lastFailureTime time.Time
	mutex           sync.RWMutex
	
	// Configuration
	failureThreshold int
	timeout          time.Duration
	successThreshold int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout time.Duration, successThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		successThreshold: successThreshold,
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) >= cb.timeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// OnSuccess records a successful execution
func (cb *CircuitBreaker) OnSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.state {
	case StateClosed:
		cb.failureCount = 0
	case StateHalfOpen:
		cb.failureCount = 0
		cb.state = StateClosed
	}
}

// OnFailure records a failed execution
func (cb *CircuitBreaker) OnFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	} else if cb.state == StateHalfOpen {
		cb.state = StateOpen
	}
}

// Execute wraps a function with circuit breaker logic
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.CanExecute() {
		return NewAnalysisError(ErrCodeNetworkError, "Circuit breaker is open")
	}
	
	err := fn()
	if err != nil {
		cb.OnFailure()
	} else {
		cb.OnSuccess()
	}
	
	return err
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.state = StateClosed
	cb.failureCount = 0
}
