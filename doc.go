// Package mockconn provides mock object of net.Conn.
//
// You can use it to test socket communication.
//
// Package Usage
//
// mockconn.New() returns mock socket instance that satisfies net.Conn interface.
// You can pass it as actual socket in your test code.
//
// If you passed t *testing.T instance to mockconn.New(),
// the socket shows error when the socket is not used in expected scenario.
//
// You can specify expected socket's communication pattern by using Read()/Write()/Close() functions.
//
//   conn := mockconn.New(t)
//   conn.SetExpectedActions(
//       mockconn.Read([]byte("sunmontue"),
//       mockconn.Write([]byte("ok"),
//       mockconn.Close(),
//   )
//
//   // use the socket in your production code
//
//   conn.Verify()
//
// Even if there are several uncalled actions, calling Verify() method detects it and dump detailed report
//
// https://www.flickr.com/photos/shibukawa/24644611414/
//
// Restrictions
//
// Now timeout functions are not implemented. You can test timeout scenarios.
package mockconn
