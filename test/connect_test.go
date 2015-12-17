package test

import "testing"

// func ConnectTCPTest(t *testing.T) {
// 	s := NewTCPServerHelper(t)
// 	defer s.Close()
// 	c := s.GetTCPClient()
// 	defer c.Close()
// }

func ConnectTLSTest(t *testing.T) {
	s := NewTLSServerHelper(t)
	defer s.Close()
	c := s.GetTLSClient(true)
	defer c.Close()
}
