package main

import "testing"

// Test if the function recognize 192.168.0.1-255 notation
func TestParseIPDash(t *testing.T) {
	errNotation := "192.168.0.1"
	dashedNotation := "192.168.0.1-2"
	_, err := parseIPRange(errNotation)

	if err == nil {
		t.Fatalf("The function do not throw error with %s notation (the correct notation express a range of IPs)", errNotation)
	}

	ipList, err := parseIPRange(dashedNotation)

	if err != nil {
		t.Fatalf("%s", err)
	}

	if len(ipList) != 2 {
		t.Fatalf("len(ipList): expected 2 but is %d", len(ipList))
	}

	if (ipList[0] != "192.168.0.1" || ipList[1] != "192.168.0.2") {
		t.Fatalf("parseIPDash func not works")
	}
}