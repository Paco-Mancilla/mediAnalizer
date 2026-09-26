package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

const keyServerAddr = "serverAddr"

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type helloHandler struct {
	db *sql.DB
}

func (h *helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, from a handler in a DB connection!\n")
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s got / resquest\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "This is my Website, and It is working!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s got /hello request\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

	ctx, cancelCtx := context.WithCancel(context.Background())

	serverOne := &http.Server{
		Addr:    ":3333",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	serverTwo := &http.Server{
		Addr:    ":4444",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	go func() {
		err := serverOne.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server one is closed\n")
		} else if err != nil {
			fmt.Printf("Error listening for server one: %s \n", err)
		}
		cancelCtx()
	}()

	go func() {
		err := serverTwo.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server two is closed\n")
		} else if err != nil {
			fmt.Printf("Error listening for server two: %s \n", err)
		}
		cancelCtx()
	}()
	<-ctx.Done()

	err := http.ListenAndServe(":3333", mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
