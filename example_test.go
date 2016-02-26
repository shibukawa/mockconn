package mockconn_test

import (
	"fmt"
	"github.com/shibukawa/mockconn"
)

func Example() {
	mock := mockconn.New(nil)
	mock.SetExpectedActions(
		mockconn.Read([]byte("welcome from server")),
		mockconn.Write([]byte("hello!!")),
		mockconn.Close(),
	)
	buffer := make([]byte, 100)
	n, _ := mock.Read(buffer)
	fmt.Println(string(buffer[:n]))
	// Output:
	// welcome from server
	mock.Write([]byte("hello!!"))
	mock.Close()
}

func ExampleConn_Verify() {
	mock := mockconn.New(nil)
	mock.SetExpectedActions(
		mockconn.Write([]byte("hello!!")),
		mockconn.Close(),
	)
	mock.Write([]byte("good morning!!"))

	for _, err := range mock.Verify() {
		fmt.Println(err)
	}
	// Output:
	// Error: socket scenario 1 - Write() expected="hello!!" actual="good morning!!"
	// Error: mock socket scenario 1 - there is remained data to write: "hello!!"
	// Error: Unconsumed senario exists - 2/2
}
