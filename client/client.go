package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	time "github.com/frederikgantriis/grpcGolang/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var T int32

func main() {
	if len(os.Args) != 3 {
		log.Printf("Please run the client with an URL and a username")
		return
	}

	T = 0

	waitc := make(chan struct{})
	conn, _ := grpc.Dial(os.Args[1], grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	client := time.NewChittyChatClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Joining the chat
	stream, _ := client.Chat(ctx)
	T++
	stream.Send(&chat.Message{Username: os.Args[2], T: T})

	//Recieve messages
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			// Update timestamp
			T = Max(T, in.GetT()) + 1
			// Log message
			log.Println(in.Username+": "+in.Msg, "Lamport:", T)
		}
	}()

	// Send messages
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		if len(msg) > 128 {
			log.Println("Messages can be at most 128 characters long")
			continue
		}
		T++
		err := stream.Send(&chat.Message{Username: os.Args[2], Msg: msg, T: T})

		if err != nil {
			panic(err)
		}

	}

	<-waitc
}

func Max(i int32, j int32) int32 {
	if i > j {
		return i
	}

	return j
}
