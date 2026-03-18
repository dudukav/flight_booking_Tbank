package client

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type State string

const (
    CLOSED   State = "CLOSED"    
    OPEN     State = "OPEN"    
    HALF_OPEN State = "HALF_OPEN"
)

type CircuitBreaker struct {
    name           string
    state          State
    failures       int
    threshold      int
    timeout        time.Duration
    lastFailure    time.Time
    mu             sync.RWMutex
}

func circuitBreakerInterceptor(cb *CircuitBreaker) grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, 
                cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        
        cb.mu.RLock()
        state := cb.state
        cb.mu.RUnlock()
        
        switch state {
        case OPEN:
            cb.mu.RLock()
            if time.Since(cb.lastFailure) > cb.timeout {
                cb.mu.RUnlock()
                cb.mu.Lock()
                cb.state = HALF_OPEN
                cb.mu.Unlock()
                log.Printf("[%s] OPEN → HALF_OPEN", cb.name)
            } else {
                cb.mu.RUnlock()
                return status.Error(codes.Unavailable, "circuit breaker OPEN")
            }
            
        case HALF_OPEN:
            err := invoker(ctx, method, req, reply, cc, opts...)
            cb.mu.Lock()
            defer cb.mu.Unlock()
            
            if err != nil {
                cb.onFailure()
                log.Printf("[%s] HALF_OPEN → OPEN (probe failed)", cb.name)
                return err
            }
            cb.onSuccess()
            log.Printf("[%s] HALF_OPEN → CLOSED (probe success)", cb.name)
            return nil
            
        case CLOSED:
            err := invoker(ctx, method, req, reply, cc, opts...)
            if err != nil {
                cb.onFailure()
            } else {
                cb.onSuccess()
            }
            return err
        }
        return nil
    }
}

func (cb *CircuitBreaker) onFailure() {
    cb.failures++
    cb.lastFailure = time.Now()
    if cb.failures >= cb.threshold {
        cb.state = OPEN
        log.Printf("[%s] CLOSED → OPEN (%d/%d failures)", cb.name, cb.failures, cb.threshold)
    }
}

func (cb *CircuitBreaker) onSuccess() {
    cb.failures = 0
    if cb.state == HALF_OPEN {
        cb.state = CLOSED
    }
}