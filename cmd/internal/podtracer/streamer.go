package podtracer

import (
	"context"
	"io"
	"log"

	pb "github.com/fennec-project/snoopy-operator/endpoint/proto"

	"google.golang.org/grpc"
)

type Streamer struct {
	client pb.DataEndpointClient
}

func (s *Streamer) Init(ip string, port string) error {

	// dial server
	conn, err := grpc.Dial(ip+":"+port, grpc.WithInsecure())
	if err != nil {
		log.Print(err.Error())
		return err
	}

	// setup client
	s.client = pb.NewDataEndpointClient(conn)

	return nil
}
func (s Streamer) Write(p []byte) (n int, err error) {

	// create stream
	stream, err := s.client.ExportPodData(context.Background())
	if err != nil {
		log.Print(err.Error())
		return 0, err
	}

	ctx := stream.Context()
	done := make(chan bool)

	// first goroutine sends poddata to snoopy data endpoint
	go func() {

		pd := pb.PodData{Name: "podtest", Data: p}
		if err := stream.Send(&pd); err != nil {
			log.Fatalf("can not send %v", err)
		}

		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	// second goroutine receives response stream from data endpoint
	// if stream is finished it closes done channel
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			message := resp.Message
			log.Printf("Received: %s", message)
		}
	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		_, ok := (<-done)
		if ok {
			close(done)
		}
	}()

	<-done
	log.Print("finished with pod data")
	return len(p), nil
}
