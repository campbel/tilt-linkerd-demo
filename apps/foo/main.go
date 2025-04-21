package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "github.com/campbel/tilt-linkerd-demo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	barHTTPURL = getEnv("BAR_HTTP_URL", "bar:80")
	bazHTTPURL = getEnv("BAZ_HTTP_URL", "baz:80")
	barGRPCURL = getEnv("BAR_GRPC_URL", "bar:9090")
	bazGRPCURL = getEnv("BAZ_GRPC_URL", "baz:9090")
	useGRPC = getEnv("USE_GRPC", "false") == "true"
)

func main() {
	// Start gRPC server
	go startGRPCServer()
	
	// Start HTTP server
	http.ListenAndServe(":8080", http.HandlerFunc(handler))
}

func handler(w http.ResponseWriter, r *http.Request) {
	barChan, bazChan := make(chan Response), make(chan Response)
	
	go func() {
		if useGRPC {
			barChan <- getBarGRPC()
		} else {
			barChan <- getBar()
		}
	}()
	
	go func() {
		if useGRPC {
			bazChan <- getBazGRPC()
		} else {
			bazChan <- getBaz()
		}
	}()

	barResponse, bazResponse := <-barChan, <-bazChan

	if barResponse.Status != http.StatusOK || bazResponse.Status != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "%s\n%s", barResponse, bazResponse)
}

func getResource(url string) Response {
	resp, err := http.Get("http://" + url)
	if err != nil {
		return Response{
			URL:   url,
			Error: err,
		}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{
			URL:   url,
			Error: err,
		}
	}
	return Response{
		Status: resp.StatusCode,
		URL:    url,
		Body:   string(body),
	}
}

func getBar() Response {
	return getResource(barHTTPURL)
}

func getBaz() Response {
	return getResource(bazHTTPURL)
}

func getBarGRPC() Response {
	conn, err := grpc.Dial(barGRPCURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Response{
			URL:   barGRPCURL,
			Error: err,
		}
	}
	defer conn.Close()
	
	client := pb.NewBarClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	
	resp, err := client.GetInfo(ctx, &pb.InfoRequest{
		Client: "foo",
		Headers: map[string]string{
			"User-Agent": "foo-grpc-client",
		},
	})
	
	if err != nil {
		return Response{
			URL:   barGRPCURL,
			Error: err,
		}
	}
	
	return Response{
		Status: int(resp.Status),
		URL:    barGRPCURL,
		Body:   resp.Message,
	}
}

func getBazGRPC() Response {
	conn, err := grpc.Dial(bazGRPCURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Response{
			URL:   bazGRPCURL,
			Error: err,
		}
	}
	defer conn.Close()
	
	client := pb.NewBazClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	
	resp, err := client.GetInfo(ctx, &pb.InfoRequest{
		Client: "foo",
		Headers: map[string]string{
			"User-Agent": "foo-grpc-client",
		},
	})
	
	if err != nil {
		return Response{
			URL:   bazGRPCURL,
			Error: err,
		}
	}
	
	return Response{
		Status: int(resp.Status),
		URL:    bazGRPCURL,
		Body:   resp.Message,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type Response struct {
	Status int
	URL    string
	Body   string
	Error  error
}

func (r Response) String() string {
	return fmt.Sprintf(`
---
   url: %s
status: %d
  body: %s
 error: %v
	`, r.URL, r.Status, r.Body, r.Error)
}

// gRPC server implementation
type fooServer struct {
	pb.UnimplementedFooServer
}

func (s *fooServer) GetInfo(ctx context.Context, req *pb.InfoRequest) (*pb.InfoResponse, error) {
	hostname, _ := os.Hostname()
	
	return &pb.InfoResponse{
		Message:  "Hello from foo gRPC server!",
		Hostname: hostname,
		Headers:  req.Headers,
		Status:   200,
	}, nil
}

func startGRPCServer() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	
	s := grpc.NewServer()
	pb.RegisterFooServer(s, &fooServer{})
	
	log.Println("Starting gRPC server on :9090")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}