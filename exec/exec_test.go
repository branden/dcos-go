package exec

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ce, err := Run(ctx, "ls", []string{"-la"})
	if err != nil {
		t.Fatal(err)
	}

	buffer := new(bytes.Buffer)
	io.Copy(buffer, ce)
	err = <-ce.Done
	if err != nil {
		t.Fatalf("Return should be nil. Got %s", err)
	}

	debugOutput := buffer.String()
	scanner := bufio.NewScanner(buffer)
	var foundString int
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "exec.go") || strings.Contains(scanner.Text(), "exec_test.go") {
			foundString++
		}
	}

	if foundString != 2 {
		t.Fatalf("Expecting `exec.go` and `exec_test.go` in output. Got: %s", debugOutput)
	}
}

func TestRunTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ce, err := Run(ctx, "bash", []string{"./fixture/infinite.sh"})
	if err != nil {
		t.Fatal(err)
	}
	buffer := new(bytes.Buffer)
	io.Copy(buffer, ce)
	err = <-ce.Done
	if err != context.DeadlineExceeded {
		t.Fatalf("Return should be %s. Got %s", context.DeadlineExceeded, err)
	}
}

func TestRunCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	ce, err := Run(ctx, "bash", []string{"./fixture/infinite.sh"})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	buffer := new(bytes.Buffer)
	io.Copy(buffer, ce)
	err = <-ce.Done
	if err != context.Canceled {
		t.Fatalf("Expected %s .Got %s", context.Canceled, err)
	}
}

func TestBadReturnCode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ce, err := Run(ctx, "command_no_found", []string{"abc"})
	if err != nil {
		t.Fatal(err)
	}
	buffer := new(bytes.Buffer)
	io.Copy(buffer, ce)
	err = <-ce.Done
	if !strings.Contains(err.Error(), "command_no_found") {
		t.Fatalf("Expected `command_no_found` in error output.Got: %s", err)
	}
}

func TestOutput(t *testing.T) {
	stdout, stderr, code, err := Output(nil, "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}

	stdoutStr := string(stdout)
	stderrStr := string(stderr)

	if stdoutStr != "hello\n" {
		t.Fatalf("expect output hello. Got %s", stdoutStr)
	}

	if stderrStr != "" {
		t.Fatalf("expect empty stderr. Got %s", stderrStr)
	}

	if code != 0 {
		t.Fatalf("expect exit code 0. Got %d", code)
	}

	// Expect error
	stdout, stderr, code, err = Output(nil, "ec", "hello")
	if err == nil {
		t.Fatal("expect error got nil")
	}

	stdoutStr = string(stdout)
	stderrStr = string(stderr)

	if stdoutStr != "" || stderrStr != "" {
		t.Fatalf("stdout and stderr must be empty. Got %s, %s", stdoutStr, stderrStr)
	}

	if code != 0 {
		t.Fatalf("expect exit code 0. Got %d", code)
	}
}

func TestOutputTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*100)
	defer cancel()
	_, _, _, err := Output(ctx, "sleep", "10")
	if err == nil {
		t.Fatal("expect error got nil")
	}
}

func TestOutputTimeoutPass(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, _, _, err := Output(ctx, "sleep", "1")
	if err != nil {
		t.Fatalf("expect nil error. Got %s", err)
	}
}

func TestReturnCode(t *testing.T) {
	_, _, code, err := Output(context.TODO(), "/bin/bash", "./fixture/return-err.sh")
	if err != nil {
		t.Fatalf("expect nil error. Got %s", err)
	}

	if code != 10 {
		t.Fatalf("expect return code 10. Got %d", code)
	}
}
