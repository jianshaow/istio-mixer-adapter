package main

import (
	"fmt"
	"os"

	authzadapter "github.com/jianshaow/istio-mixer-adapter/adapter/authzadapter"
)

func main() {
	addr := ""
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	s, err := authzadapter.NewAuthzAdapter(addr)
	if err != nil {
		fmt.Printf("unable to start server: %v", err)
		os.Exit(-1)
	}

	shutdown := make(chan error, 1)
	go func() {
		s.Run(shutdown)
	}()
	_ = <-shutdown
}
