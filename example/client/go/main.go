package main

import (
	"context"
	"io"
	"log"
	"os"
	"sync"

	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:4770", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewGripmockClient(conn)

	// Contact the server and print out its response.
	name := "tokopedia"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r, err := c.SayHello(context.Background(), &Request{Name: name})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go serverStream(c, wg)

	wg.Add(1)
	go clientStream(c, wg)

	wg.Wait()
}

// server to client streaming
func serverStream(c GripmockClient, wg *sync.WaitGroup) {
	defer wg.Done()
	req := &Request{
		Name: "server-to-client-streaming",
	}
	stream, err := c.ServerStream(context.Background(), req)
	if err != nil {
		log.Fatal("server stream error: %v", err)
	}

	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal("s2c error: %v", err)
		}

		log.Printf("s2c message: %s\n", reply.Message)
	}
}

// client to server streaming
func clientStream(c GripmockClient, wg *sync.WaitGroup) {
	defer wg.Done()
	stream, err := c.ClientStream(context.Background())
	if err != nil {
		log.Fatalf("c2s error: %v", err)
	}

	requests := []Request{
		{
			Name: "c2s-1",
		}, {
			Name: "c2s-2",
		},
	}
	for _, req := range requests {
		err := stream.Send(&req)
		if err != nil {
			log.Fatalf("c2s error: %v", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("c2s error: %v", err)
	}

	log.Printf("c2s message: %v", reply.Message)
}
