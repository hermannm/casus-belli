package tests

import (
	"immerse/hermannia/server/hello"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	var result string = hello.HelloWorld()
	var correct string = "hello, world"

	if result != correct {
		t.Errorf("Hello World was incorrect, got: %s, want %s.", result, correct)
	}
}
