package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const spinBinary = "spin"

func retryGet(t *testing.T, url string) *http.Response {
	t.Helper()

	const maxTries = 600 // (10min)
	for i := 1; i < maxTries; i++ {
		// Catch call to `Fail` in other goroutine
		if t.Failed() {
			t.FailNow()
		}
		if res, err := http.Get(url); err != nil {
			t.Log(err)
		} else {
			return res
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Get request timeout: ", url)
	return nil
}

type testSpin struct {
	cancel func()
	url    string
	cmd    *exec.Cmd
}

func startSpin(t *testing.T, dir string) *testSpin {
	buildApp(t, dir)

	url := getFreePort(t)

	// long timeout because... ci
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	cmd := exec.CommandContext(ctx, spinBinary, "up", "--listen", url)
	cmd.Dir = dir
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		t.Log(stderr.String())
		t.Fatal(err)
	}

	go func() {
		_ = cmd.Wait()
		if ctx.Err() == nil {
			t.Log("spin exited before the test finished:", cmd.ProcessState)
			t.Log("stderr:\n", stderr.String())
			t.Fail()
		}
	}()

	return &testSpin{
		cancel: cancel,
		url:    fmt.Sprintf("http://%s", url),
		cmd:    cmd,
	}
}

func buildApp(t *testing.T, dir string) {
	t.Helper()

	t.Log("building application:", dir)

	cmd := exec.Command(spinBinary, "build")
	cmd.Dir = dir

	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		t.Log(stderr.String())
		t.Errorf("Failed to build %q, %v", dir, err)
	}
}

func TestHTTPTriger(t *testing.T) {
	spin := startSpin(t, "http/testdata/http-tinygo")
	defer spin.cancel()

	resp := retryGet(t, spin.url+"/hello")
	spin.cancel()
	if resp.Body == nil {
		t.Fatal("body is nil")
	}
	t.Log(resp.Status)
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// assert response body
	want := "== RESPONSE ==\nHello spinframework!\nHello again spinframework!\n"
	got := string(b)
	if want != got {
		t.Fatalf("body is not equal: want = %q got = %q", want, got)
	}

	// assert response header
	if resp.Header.Get("foo") != "bar" {
		t.Fatal("header 'foo' was not set")
	}
}

// TestBuildExamples ensures that the tinygo examples will build successfully.
func TestBuildExamples(t *testing.T) {
	examples, err := os.ReadDir("examples")
	if err != nil {
		t.Fatal(err)
	}
	for _, example := range examples {
		example := example
		t.Run(example.Name(), func(t *testing.T) {
			t.Parallel()
			buildApp(t, filepath.Join("examples", example.Name()))
		})
	}
}

func getFreePort(t *testing.T) string {
	t.Helper()

	a, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("failed to get free port: ", err)
	}

	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		t.Fatal("failed to get free port: ", err)
	}
	l.Close()
	return l.Addr().String()
}
