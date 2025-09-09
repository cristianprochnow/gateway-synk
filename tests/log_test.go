package tests

import (
	"bytes"
	"io"
	"os"
	"strings"
	"synk/gateway/app/util"
	"testing"
)

func TestLog(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	message := "This is a test log message"
	util.Log(message)

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	os.Stdout = originalStdout

	got := buf.String()

	if got == "" {
		t.Errorf("util.Log() print empty content")
	}
}

func TestLogRoute(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	route := "/api/v1/users"
	message := "Request received"
	util.LogRoute(route, message)

	w.Close()

	var capturedOutput strings.Builder
	io.Copy(&capturedOutput, r)

	os.Stdout = originalStdout

	got := capturedOutput.String()

	if got == "" {
		t.Errorf("util.LogRoute() print empty content")
	}
}
