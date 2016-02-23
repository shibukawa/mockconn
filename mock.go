package mockconn

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// ActionType is an enum value to identify actions in scenario
type ActionType int

const (
	// ReadAction type
	ReadActionType ActionType = iota
	// WriteAction type
	WriteActionType
	// CloseAction type
	CloseActionType
	nullActionType
)

// Action is an interface of actions in scenario
type Action interface {
	Type() ActionType
}

type readAction struct {
	data []byte
}

// Read creates action to read
//
// You can read data by using several Conn.Read() call from single Read action:
//
//   conn := mockconn.New(t)
//   conn.SetExpectedActions(
//       Read([]byte("sunmontue")
//   )
//   d := make([]byte, 3)
//   conn.Read(d) // 'sun' : ok
//   conn.Read(d) // 'mon' : ok
//   conn.Read(d) // 'tus' : ok
//   conn.Read(d) // ''    : ok
func Read(data []byte) Action {
	return &readAction{
		data: data,
	}
}

func (r readAction) Type() ActionType {
	return ReadActionType
}

type writeAction struct {
	data []byte
}

// Write creates action to write
//
// You can write data by using several Conn.Write() call to single Write action:
//
//   conn := mockconn.New(t)
//   conn.SetExpectedActions(
//       Write([]byte("sunmontue")
//   )
//   conn.Write([]byte("sun") // ok
//   conn.Write([]byte("mon") // ok
//   conn.Write([]byte("tue") // ok
func Write(data []byte) Action {
	return &writeAction{
		data: data,
	}
}

func (w writeAction) Type() ActionType {
	return WriteActionType
}

type closeAction struct {
}

// Close creates action to close
func Close() Action {
	return &closeAction{}
}

func (c closeAction) Type() ActionType {
	return CloseActionType
}

type nullAction struct {
}

func (n nullAction) Type() ActionType {
	return nullActionType
}

// Conn is a mock object that has net.Conn interface
type Conn struct {
	t          *testing.T
	errors     []error
	scenario   []Action
	current    int
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (c Conn) getAction(i int) Action {
	if i < len(c.scenario) {
		return c.scenario[i]
	}
	return &nullAction{}
}

// NewConn creates mock connection instance
// if t is passed, it calls t.Errorf in unit tests.
func New(t *testing.T) *Conn {
	return &Conn{
		t:          t,
		localAddr:  &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345},
		remoteAddr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080},
	}
}

// SetLocalAddr sets local address to return
func (c *Conn) SetLocalAddr(addr net.Addr) {
	c.localAddr = addr
}

// SetRemoteAddr sets remote address to return
func (c *Conn) SetRemoteAddr(addr net.Addr) {
	c.remoteAddr = addr
}

// SetExpectedActions sets expected behavior
func (c *Conn) SetExpectedActions(scenario ...Action) {
	c.scenario = scenario
}

// Verify returns all errors
func (c *Conn) Verify() []error {
	errors := make([]error, len(c.errors))
	copy(errors, c.errors)
	current := c.getAction(c.current)
	switch current.Type() {
	case ReadActionType:
		read := current.(*readAction)
		if len(read.data) > 0 {
			c.addError(fmt.Errorf("mock socket scenario %d: there is remained data to read: '%s'", c.current+1, string(read.data)))
		}
		c.current++
	case WriteActionType:
		write := current.(*writeAction)
		c.addError(fmt.Errorf("mock socket scenario %d: there is remained data to write: '%s'", c.current+1, string(write.data)))
		c.current++
	}
	for i := c.current; i < len(c.scenario); i++ {
		current := c.getAction(i)
		switch current.Type() {
		case ReadActionType:
			c.addError(fmt.Errorf("mock socket scenario %d: Read() should be called", c.current+1))
		case WriteActionType:
			c.addError(fmt.Errorf("mock socket scenario %d: Write() should be called", c.current+1))
		case CloseActionType:
			c.addError(fmt.Errorf("mock socket scenario %d: Close() should be called", c.current+1))
		}
	}
	result := c.errors
	c.errors = errors
	return result
}

func (c *Conn) addError(err error) error {
	c.errors = append(c.errors, err)
	if c.t != nil {
		c.t.Error(err.Error())
	}
	return err
}

// Read reads data from the connection.
// Read can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (n int, err error) {
	current := c.getAction(c.current)
	switch current.Type() {
	case ReadActionType:
		read := current.(*readAction)
		if len(read.data) > 0 {
			n := copy(b, read.data)
			read.data = read.data[n:]
			return n, nil
		}
		next := c.getAction(c.current + 1)
		if next.Type() == ReadActionType {
			c.current++
			return c.Read(b)
		}
		return 0, nil
	case WriteActionType:
		write := current.(*writeAction)
		if len(write.data) == 0 {
			c.current++
			current = c.scenario[c.current]
		} else {
			return 0, c.addError(fmt.Errorf("socket scenario %d: should close, but Read() is called", c.current+1))
		}
	case CloseActionType:
		return 0, c.addError(fmt.Errorf("socket scenario %d: should close, but Read() is called", c.current+1))
	}
	return 0, nil
}

// Write writes data to the connection.
// Write can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (n int, err error) {
	current := c.getAction(c.current)
	switch current.Type() {
	case ReadActionType:
		read := current.(*readAction)
		if len(read.data) > 0 {
			return 0, c.addError(fmt.Errorf("socket scenario %d: should read data, but Write() is called", c.current+1))
		}
		c.current++
		return c.Write(b)
	case WriteActionType:
		write := current.(*writeAction)
		if len(b) <= len(write.data) {
			same := true
			for i, ch := range b {
				if ch != write.data[i] {
					same = false
					break
				}
			}
			if !same {
				return 0, c.addError(fmt.Errorf("socket scenario %d: Write() expected='%s' actual='%s'", c.current+1, string(write.data), string(b)))
			}
			if len(b) == len(write.data) {
				c.current++
			} else {
				write.data = write.data[len(b):]
			}
			return len(b), nil
		}
		return 0, c.addError(fmt.Errorf("socket scenario %d: Write() expected='%s' actual='%s'", c.current+1, string(write.data), string(b)))
	case CloseActionType:
		return 0, c.addError(fmt.Errorf("socket scenario %d: should close, but Write() is called", c.current+1))
	}
	return 0, nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	current := c.getAction(c.current)
	switch current.Type() {
	case ReadActionType:
		read := current.(*readAction)
		if len(read.data) > 0 {
			return c.addError(fmt.Errorf("socket scenario %d: should read data, but Close() is called", c.current+1))
		}
		c.current++
		return c.Close()
	case WriteActionType:
		return c.addError(fmt.Errorf("socket scenario %d: should write data, but Close() is called", c.current+1))
	case CloseActionType:
		c.current++
	}
	return nil
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.localAddr
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future I/O, not just
// the immediately following call to Read or Write.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}
