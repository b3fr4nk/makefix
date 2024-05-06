package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/joho/godotenv"
)

func TestTableIsMarkdownFile(t * testing.T) {
	var tests = []struct {
		input string
		expected bool
	}{
		{"test.md", true},
		{"test.txt", false},
		{"md.txt", false},
	}
	for _, test := range tests {
		if output := isMarkdownFile(test.input); output != test.expected {
			t.Error("Test Failed: {} input, {} expected, received: {}", test.input, test.expected, output)
		}
	}
}

func TestTableGetMarkdown(t *testing.T) {
	err := godotenv.Load()
  	if err != nil {
		fmt.Print("error")
    	log.Fatal("Error loading .env file")
  	}
	files := []string{"Readme.md", "test2.md"}
	var tests = []struct {
		input string
		expected []string
	}{
		{"https://github.com/b3fr4nk/test-makefix", files},
	}

	for _, test := range tests {
		_, output, _ := getMarkdown(test.input)

		for i := range test.expected {
			if len(output) < i {
				t.Error("Test Failed: test.input test.expected, {} expected, received: {}", test.input, len(test.expected), len(output))
			}
			if output[i] != test.expected[i] {
				t.Error("Test Failed: {} input, {} expected, received: {}", test.input, test.expected[i], output[i])
			}
		}
	}
}