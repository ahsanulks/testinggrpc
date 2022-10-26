package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ahsanulks/protofile"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var clientService protofile.AddServiceClient

func main() {
	conn, err := grpc.Dial(":4040", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	clientService = protofile.NewAddServiceClient(conn)
	// countDown(clientService, 10)
	// multipleSum(clientService)
	realTimeMultipleSum(clientService)

	r := mux.NewRouter()
	r.HandleFunc("/add", AddHandler)
	r.HandleFunc("/multiple", MultipleHandler).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]int64)

	json.NewDecoder(r.Body).Decode(&body)
	x, y := body["x"], body["y"]

	req := protofile.Request{X: x, Y: y}
	response, err := clientService.Add(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"total": response.Result})
}

func MultipleHandler(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]int64)

	json.NewDecoder(r.Body).Decode(&body)
	x, y := body["x"], body["y"]

	req := protofile.Request{X: x, Y: y}
	response, err := clientService.Multiply(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"total": response.Result})
}

func countDown(client protofile.AddServiceClient, number int) {
	log.Println("contdown number", number)

	stream, err := client.CountDown(context.Background(), &protofile.CountDownRequest{
		Number: 10,
	})
	if err != nil {
		panic("error when connect to server")
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			stream.CloseSend()
			log.Println("end receiving message")
			break
		}
		if err != nil {
			panic("panic when getting message")
		}
		log.Println("current countdown number", msg.Number)
		time.Sleep(1 * time.Second)
	}
}

func multipleSum(client protofile.AddServiceClient) {
	numbers := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	stream, err := client.MultipleSum(context.Background())
	if err != nil {
		panic("error when connect to server")
	}

	for _, number := range numbers {
		log.Println("send request to server with number", number)
		stream.Send(&protofile.MultipleSumRequest{
			Number: number,
		})
		time.Sleep(1 * time.Second)
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		panic("error when closing connection to server")
	}
	log.Println("get response from server is", resp.Result)
}

func realTimeMultipleSum(client protofile.AddServiceClient) {
	numbers := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	stream, err := client.RealtimeMultipleSum(context.Background())
	if err != nil {
		panic("error when connect to server")
	}

	waitc := make(chan struct{})
	go func() {
		for _, number := range numbers {
			log.Println("send request to server with number", number)
			stream.Send(&protofile.RealtimeMultipleSumRequest{
				Number: number,
			})
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				panic("error when get message from server")
			}
			log.Println("current multiple sum is", res.Number)
		}
		close(waitc)
	}()

	<-waitc
}
