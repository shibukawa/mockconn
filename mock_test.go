package mockconn

import (
	"net"
	"testing"
)

func TestAssignToNetConn(t *testing.T) {
	var mock net.Conn = New(t)
	mock.Close()
}

func TestReadCloseScenario(t *testing.T) {
	mock := New(t)
	mock.SetExpectedActions(
		Read([]byte("hello world")),
		Close(),
	)
	buffer := make([]byte, 100)
	n, err := mock.Read(buffer)
	if n != len("hello world") {
		t.Errorf("Read result: %d", n)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	n, err = mock.Read(buffer) // reading from empty socket is valid
	if n != len("") {
		t.Errorf("Read result: %d", n)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	mock.Close()
	errors := mock.Verify()
	if len(errors) != 0 {
		t.Errorf("mock shouldn't have any errors, but %d", len(errors))
		for i, err := range errors {
			t.Log(i+1, err.Error())
		}
	}
}

func TestReadError(t *testing.T) {
	mock := New(nil)
	mock.SetExpectedActions(
		Close(),
	)
	buffer := make([]byte, 100)
	_, err := mock.Read(buffer)
	if err == nil {
		t.Error("err should not be nil")
	}
	errors := mock.Verify()
	if len(errors) != 2 {
		t.Errorf("mock should have 2 errors, but %d", len(errors))
		for i, err := range errors {
			t.Log(i+1, err.Error())
		}
	}
}

func TestCloseError(t *testing.T) {
	mock := New(nil)
	mock.SetExpectedActions(
		Read([]byte("hello world")),
	)
	err := mock.Close()
	if err == nil {
		t.Error("err should not be nil")
	}
	errors := mock.Verify()
	if len(errors) != 3 {
		t.Errorf("mock should have 2 errors, but %d", len(errors))
		for i, err := range errors {
			t.Log(i+1, err.Error())
		}
	}
}

func TestWriteCloseScenario(t *testing.T) {
	mock := New(t)
	mock.SetExpectedActions(
		Write([]byte("helloworld")),
		Close(),
	)
	n, err := mock.Write([]byte("hello"))
	if n != len("hello") {
		t.Errorf("Write result: %d", n)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	n, err = mock.Write([]byte("world"))
	if n != len("world") {
		t.Errorf("Write result: %d", n)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	mock.Close()
	errors := mock.Verify()
	if len(errors) != 0 {
		t.Errorf("mock shouldn't have any errors, but %d", len(errors))
		for i, err := range errors {
			t.Log(i+1, err.Error())
		}
	}
}

func TestWriteError(t *testing.T) {
	mock := New(nil)
	mock.SetExpectedActions(
		Close(),
	)
	_, err := mock.Write([]byte("hello world"))
	if err == nil {
		t.Error("err should not be nil")
	}
	errors := mock.Verify()
	if len(errors) != 2 {
		t.Errorf("mock should have 2 errors, but %d", len(errors))
		for i, err := range errors {
			t.Log(i+1, err.Error())
		}
	}
}
