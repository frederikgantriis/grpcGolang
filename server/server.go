package main

import (
	"io"
	"log"
	"net"
	"os"

	time "github.com/frederikgantriis/grpcGolang/chat"

	"google.golang.org/grpc"
)

type Server struct {
	time.UnimplementedChittyChatServer
	streams []time.ChittyChat_ChatServer
}

func (s *Server) connect(newStream chat.ChittyChat_ChatServer) string {
	s.streams = append(s.streams, newStream)
	// Notify all clients that a new client has joined
	nameMsg, _ := newStream.Recv()

	for _, client := range s.streams {
		client.SendMsg(&chitty_chat.Message{Username: "Server", Msg: nameMsg.GetUsername() + " has joined the chat", T: nameMsg.GetT()})
	}
	log.Println(nameMsg.GetUsername(), "has joined the chat", "Lamport:", nameMsg.GetT())
	return nameMsg.GetUsername()
}

func (s *Server) Chat(stream chitty_chat.ChittyChat_ChatServer) error {
	var user string
	user = s.connect(stream)

	go func() {
		for {
			msg, err := stream.Recv()

			if err == io.EOF {
				return
			} else if err != nil {
				// Log on server
				log.Println(user, "left the chat")
				// Remove stream from server client list
				remove(s.streams, stream)
				// Notify clients
				for _, client := range s.streams {
					client.SendMsg(&chitty_chat.Message{Msg: user + " left the chat", T: msg.GetT()})
				}
				return
			}
			for _, client := range s.streams {
				client.SendMsg(&chitty_chat.Message{Username: msg.GetUsername(), Msg: msg.GetMsg(), T: msg.GetT()})
			}
		}
	}()

	waitc := make(chan struct{})

	<-waitc
	return nil
}

func main() {
	if len(os.Args) != 2 {
		log.Printf("Please input the port to run the server on")
	}

	lis, _ := net.Listen("tcp", "localhost:"+os.Args[1])

	grpcServer := grpc.NewServer()
	chitty_chat.RegisterChittyChatServer(grpcServer, &Server{})

	log.Printf("server listening at %v", lis.Addr())

	grpcServer.Serve(lis)
}

func remove(s []chitty_chat.ChittyChat_ChatServer, client chitty_chat.ChittyChat_ChatServer) []chitty_chat.ChittyChat_ChatServer {
	var i int
	for j, stream := range s {
		if stream == client {
			i = j
			break
		}
	}
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func Max(i int32, j int32) int32 {
	if i > j {
		return i
	}

	return j
}
