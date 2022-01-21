package main

import "testing"

func TestEmailIsValid(t *testing.T) {

	testcases := []struct {
		desc     string
		input    string
		expected bool
	}{
		{"missing domain", "j@email", false},
		{"missing name", "@email.com", false},
		{"missing @", "email.com", false},
		{"public domain too long", "j@email.coommm", false},
		{"valid domain with subdomain", "j@email.co.uk", true},
		{"is a valid public email", "j@email.ch", true},
	}

	var result bool
	for _, tc := range testcases {
		_, err := emailIsValid(tc.input)
		result = err == nil
		if result != tc.expected {
			t.Errorf("%v (%v): got %v expected %v", tc.input, tc.desc, result, tc.expected)
		}

	}
}

//  false

// "john@email.a" false

// "john@email.aabbc" false

// "@email.com" false

// "j@email.ch" true

// "j@e.co.uk" true
