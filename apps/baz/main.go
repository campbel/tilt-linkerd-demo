package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"

	pb "github.com/campbel/tilt-linkerd-demo/proto"
	"google.golang.org/grpc"
)

func main() {
	// Start gRPC server
	go startGRPCServer()
	
	// Start HTTP server
	http.ListenAndServe(":8080", http.HandlerFunc(handler))
}

func handler(w http.ResponseWriter, r *http.Request) {
	content := fmt.Sprintf(`
Pod details:
  Hostname: %s

Request details:
  Host: %s
  Path: %s
  Method: %s
  Protocol: %s
  RemoteAddr: %s
  RequestURI: %s
  Headers:
`, os.Getenv("HOSTNAME"), r.Host, r.URL.Path, r.Method, r.Proto, r.RemoteAddr, r.RequestURI)
	for _, k := range sortedHeaderKeys(r.Header) {
		content += fmt.Sprintf("    %s: %s\n", k, r.Header.Get(k))
	}

	fmt.Fprintln(w, content)
}

func sortedHeaderKeys(headers http.Header) []string {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// gRPC server implementation
type bazServer struct {
	pb.UnimplementedBazServer
}

func (s *bazServer) GetInfo(ctx context.Context, req *pb.InfoRequest) (*pb.InfoResponse, error) {
	hostname, _ := os.Hostname()
	
	// Build header string for message
	headersStr := ""
	headerKeys := make([]string, 0, len(req.Headers))
	for k := range req.Headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)
	
	for _, k := range headerKeys {
		headersStr += fmt.Sprintf("    %s: %s\n", k, req.Headers[k])
	}
	
	message := fmt.Sprintf(`
Pod details:
  Hostname: %s

gRPC Request details:
  Client: %s
  Headers:
%s`, hostname, req.Client, headersStr)
	
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
	pb.RegisterBazServer(s, &bazServer{})
	
	log.Println("Starting gRPC server on :9090")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}