package main

import (
	"context"
	"io"
	"log"
	"net"
	"time"

	"github.com/ahsanulks/protofile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	protofile.UnimplementedAddServiceServer
}

func main() {
	listen, err := net.Listen("tcp", ":4040")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()

	protofile.RegisterAddServiceServer(srv, &server{})
	reflection.Register(srv)

	log.Println("run server")
	if err := srv.Serve(listen); err != nil {
		panic(err)
	}
}

func (s *server) Add(context context.Context, request *protofile.Request) (*protofile.Response, error) {
	log.Println("have a request with params", request.X, request.Y)
	return &protofile.Response{Result: request.X + request.Y}, nil
}

func (s *server) Multiply(context context.Context, request *protofile.Request) (*protofile.Response, error) {
	log.Println("have a request with params", request.X, request.Y)
	return &protofile.Response{Result: request.X * request.Y}, nil
}

func (s *server) CountDown(req *protofile.CountDownRequest, stream protofile.AddService_CountDownServer) error {
	log.Println("have a countdown request with params", req.Number)
	number := req.Number
	var err error
	for i := number; i >= 0; i-- {
		log.Println("sending message", i)
		err = stream.Send(&protofile.CountDownResponse{
			Number: i,
		})
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return err
}

func (s *server) MultipleSum(stream protofile.AddService_MultipleSumServer) error {
	log.Println("have multiple sum")

	total := 0
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("total is", total)
			return stream.SendAndClose(&protofile.Response{
				Result: int64(total),
			})
		}
		if err != nil {
			panic("error when get request")
		}
		number := msg.Number
		log.Printf("current total %d, will sum with %d", total, number)
		total += int(number)
	}
}

func (s *server) RealtimeMultipleSum(stream protofile.AddService_RealtimeMultipleSumServer) error {
	log.Println("have real multiple sum")

	total := int64(0)
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("end connection")
			break
		}
		total += msg.Number
		time.Sleep(1 * time.Second)
		err = stream.Send(&protofile.RealtimeMultipleSumRequest{
			Number: total,
		})
		if err != nil {
			panic("error when send response to client")
		} else {
			log.Println("sending response to client with total", total)
		}
	}
	return nil
}
