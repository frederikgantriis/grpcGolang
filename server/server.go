package main

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/frederikgantriis/grpcGolang.git/chat"
	"google.golang.org/grpc"
)

type Server struct {
	chat.UnimplementedChittyChatServer
	streams []chat.ChittyChat_ChatServer
}

func (s *Server) connect(newStream chat.ChittyChatServer) string {
	s.streams = append(s.streams, newStream)
	nameMsg, _ := newStream.Recv()

	for _, client := range s.streams {
		client.SendMsg(&chat.Message{Username: "Server", Msg: nameMsg.GetUsername() + " has joined the chat", T: nameMsg.GetT()})
	}
	log.Println(nameMsg.GetUsername(), "has joined the chat", "Lamport:", nameMsg.GetT())
	return nameMsg.GetUsername
}

func (s *Server) Chat(stream chat.ChittyChat_ChatServer) error {
	var user string
	user = s.connect(stream)

	go func() {
		for {
			msg, err := stream.Recv()

			if err == io.EOF {
				return
			} else if err != nil {
				log.Println(user, "left the chat")
				remove(s.streams, stream)
				for _, client := range s.streams {
					client.SendMsg(&chat.Message{Msg: user + " left the chat", T: msg.GetT()})
				}
				return
			}
			for _, client := range s.streams {
				client.SendMsg(&chat.Message{Username: msg.GetUsername(), Msg: msg.GetMsg(), T: msg.GetT()})
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
	chat.RegisterChittyChatServer(grpcServer, &Server{})

	log.Printf("server listening at %v", lis.Addr())

	grpcServer.Serve(lis)
}

func remove(s []chat.ChittyChat_ChatServer, client chat.ChittyChat_ChatServer) []chat.ChittyChat_ChatServer {
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
