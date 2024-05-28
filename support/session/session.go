package session

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xhd2015/xgo/support/netutil"
)

type SessionManager interface {
	// return id and error
	Start() (string, Session, error)
	Get(id string) (Session, error)
	Destroy(id string) error
}

func NewSessionManager() SessionManager {
	return &sessionManager{}
}

type Session interface {
	SendEvents(events ...interface{}) error
	PollEvents() ([]interface{}, error)
}

type sessionImpl struct {
	ch chan interface{}
}

func (c *sessionImpl) SendEvents(events ...interface{}) error {
	for _, e := range events {
		c.ch <- e
	}
	return nil
}

func (c *sessionImpl) PollEvents() ([]interface{}, error) {
	events := c.poll(5*time.Second, 100*time.Millisecond)
	return events, nil
}

func (c *sessionImpl) poll(timeout time.Duration, pollInterval time.Duration) []interface{} {
	var events []interface{}

	timeoutCh := time.After(timeout)
	for {
		select {
		case event := <-c.ch:
			events = append(events, event)
		case <-timeoutCh:
			return events
		default:
			if len(events) > 0 {
				return events
			}
			time.Sleep(pollInterval)
		}
	}
}

type sessionManager struct {
	nextID  int64
	mapping sync.Map
}

func (c *sessionManager) Start() (string, Session, error) {
	// to avoid stale requests from older pages
	idInt := atomic.AddInt64(&c.nextID, 1)
	id := fmt.Sprintf("session_%s_%d", time.Now().Format("2006-01-02_15:04:05"), idInt)
	session := &sessionImpl{
		ch: make(chan interface{}, 100),
	}
	c.mapping.Store(id, session)
	return id, session, nil
}
func (c *sessionManager) Get(id string) (Session, error) {
	val, ok := c.mapping.Load(id)
	if !ok {
		return nil, netutil.ParamErrorf("session %s does not exist or has been removed", id)
	}
	session := val.(*sessionImpl)
	return session, nil
}

func (c *sessionManager) Destroy(id string) error {
	c.mapping.Delete(id)
	return nil
}
