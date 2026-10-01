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
	hasFirst := r.URL.Query().Has("first")
	first := r.URL.Query().Get("first")
	hasSecond := r.URL.Query().Has("second")
	second := r.URL.Query().Get("second")
	fmt.Printf("%s got / request. first(%t) = %s, second(%t) = %s\n", ctx.Value(keyServerAddr), hasFirst, first, hasSecond, second)
	io.WriteString(w, "This is my Website, and It is working!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s got /hello request\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "Hello, HTTP!\n")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Started %s %s\n", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		fmt.Printf("Completed %s %s\n", r.Method, r.URL.Path)
	})
}

func headerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-App-Version", "1.0")
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

	loggedMux := loggingMiddleware(mux)

	ctx := context.Background()

	server := &http.Server{
		Addr:    ":3333",
		Handler: loggedMux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server is closed\n")
	} else if err != nil {
		fmt.Printf("Error listening for server: %s \n", err)
	}

	<-ctx.Done()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}

}
