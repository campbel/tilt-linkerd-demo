package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"time"

	pb "github.com/campbel/tilt-linkerd-demo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	bazHTTPURL = getEnv("BAZ_HTTP_URL", "baz:80")
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
	var bazResponse string
	var err error
	
	if useGRPC {
		bazResponse, err = callBazGRPC()
	} else {
		bazResponse, err = callBazHTTP()
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	content := "Hello from bar! Here are the request details:\n\n"
	content += fmt.Sprintf("Host: %s\n", r.Host)
	content += fmt.Sprintf("Path: %s\n", r.URL.Path)
	content += fmt.Sprintf("Method: %s\n", r.Method)
	content += fmt.Sprintf("Protocol: %s\n", r.Proto)
	content += fmt.Sprintf("RemoteAddr: %s\n", r.RemoteAddr)
	content += fmt.Sprintf("RequestURI: %s\n", r.RequestURI)

	content += "Headers:\n"
	for _, k := range sortedHeaderKeys(r.Header) {
		content += fmt.Sprintf("  %s: %s\n", k, r.Header.Get(k))
	}

	content += fmt.Sprintf("Response from baz:\n%s\n", bazResponse)

	fmt.Fprintln(w, content)
}

func callBazHTTP() (string, error) {
	resp, err := http.Get("http://" + bazHTTPURL + "/foo/bar")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("status: %s\nbody: %s", resp.Status, string(body)), nil
}

func callBazGRPC() (string, error) {
	conn, err := grpc.Dial(bazGRPCURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()
	
	client := pb.NewBazClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	
	resp, err := client.GetInfo(ctx, &pb.InfoRequest{
		Client: "bar",
		Headers: map[string]string{
			"User-Agent": "bar-grpc-client",
			"Path": "/foo/bar",
		},
	})
	
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("status: %d\nbody: %s", resp.Status, resp.Message), nil
}

func sortedHeaderKeys(headers http.Header) []string {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// gRPC server implementation
type barServer struct {
	pb.UnimplementedBarServer
}

func (s *barServer) GetInfo(ctx context.Context, req *pb.InfoRequest) (*pb.InfoResponse, error) {
	var bazResponse string
	var err error
	
	if useGRPC {
		bazResponse, err = callBazGRPC()
	} else {
		bazResponse, err = callBazHTTP()
	}
	
	if err != nil {
		return nil, err
	}
	
	hostname, _ := os.Hostname()
	
	message := fmt.Sprintf("Hello from bar gRPC server!\nClient: %s\nBaz response: %s", 
		req.Client, bazResponse)
	
	return &pb.InfoResponse{
		Message:  message,
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
	pb.RegisterBarServer(s, &barServer{})
	
	log.Println("Starting gRPC server on :9090")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}