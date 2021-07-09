package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sbashilov/todo/pb"
	"github.com/sbashilov/todo/service"
	"google.golang.org/grpc"
)

const (
	sizeLen = 8
	file    = "tasks.pb"
)

func main() {
	ctx := context.Background()
	svc := &service.TaskService{}
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "require subcommand")
		os.Exit(1)
	}

	var err error
	switch cmd := flag.Arg(0); cmd {
	case "add":
		_, err = svc.Add(ctx, &pb.Task{Text: strings.Join(flag.Args()[1:], " ")})
	case "list":
		err = svc.ListPrint()
	case "done":
		err = svc.Done(flag.Arg(1))
	case "grpc":
		if err = runGrpc(svc); err != nil {
			log.Fatal(err)
		}
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func runGrpc(svc *service.TaskService) error {
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		log.Print("logged handler")
		w.Write([]byte("hello world!"))
	})
	go func() {
		log.Println("http serve start")
		http.ListenAndServe(":8080", nil)
	}()
	server := grpc.NewServer(grpc.UnaryInterceptor(
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			log.Print(info.FullMethod, req)
			return handler(ctx, req)
		},
	))
	pb.RegisterTaskServiceServer(server, svc)
	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		return errors.Wrap(err, "unable to create listener")
	}
	log.Println("grpc serve start")
	if err = server.Serve(listener); err != nil {
		return errors.Wrap(err, "unable to start server")
	}
	return nil
}
