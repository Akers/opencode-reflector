package main

import (
	"os/exec"
	"testing"
)

func TestMainCompiles(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("main.go should compile: %v\n%s", err, out)
	}
}
