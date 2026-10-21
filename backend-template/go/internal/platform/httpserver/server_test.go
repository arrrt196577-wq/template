package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
)

func TestShutdownWaitsForInFlightRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	requestContextAlive := make(chan bool, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		requestContextAlive <- r.Context().Err() == nil
		w.WriteHeader(204)
	})
	server := New(config.HTTP{ShutdownTimeout: 2 * time.Second}, handler, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- server.Serve(ctx, listener) }()
	clientResult := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_ = response.Body.Close()
		}
		clientResult <- err
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not enter handler")
	}
	cancel()
	select {
	case err := <-result:
		t.Fatalf("server exited before request finished: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	if err := <-clientResult; err != nil {
		t.Fatal(err)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if !<-requestContextAlive {
		t.Fatal("shutdown canceled the in-flight request before draining")
	}
}

func TestShutdownDeadlineForcesConnectionClosed(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	entered, finished := make(chan struct{}), make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
		close(finished)
	})
	server := New(config.HTTP{ShutdownTimeout: 10 * time.Millisecond}, handler, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- server.Serve(ctx, listener) }()
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_ = response.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not enter handler")
	}
	cancel()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("shutdown deadline was not reported")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown remained blocked")
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("forced close did not cancel active connection")
	}
	<-clientDone
}
