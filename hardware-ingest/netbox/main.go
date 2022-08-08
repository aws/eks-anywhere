package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	defer done()

	host := flag.String("host", "", "Host Server running DCIM tool")
	token := flag.String("token", "", "API token for HTTP connection with DCIM")
	tag := flag.String("tag", "eks-a", "tag for filtering devices")
	debug := flag.Bool("debug", false, "debug flag for logging")
	flag.Parse()
	if len(*host) == 0 {
		fmt.Fprintln(os.Stdout, "Host cannot be blank")
		return
	} else if len(*token) == 0 {
		fmt.Fprintln(os.Stdout, "token ID cannot be blank")
		return
	}
	if err := runClient(ctx, *host, *token, *tag, *debug); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
