package test

import "testing"

func ConnectTCPTest(t *testing.T) {
	s := NewTLSServerHelper(t)
	defer s.Close()
}

func ConnectTLSTest(t *testing.T) {
	s := NewTLSServerHelper(t)
	defer s.Close()
}
