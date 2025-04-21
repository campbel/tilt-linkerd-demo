# Building a Linkerd Demo: Adding gRPC Support (Part 2)

In [Part 1](part_1.md) of this series, we set up a simple microservices demo with Linkerd on a local Kubernetes cluster using Tilt. We created three services (foo, bar, and baz) that communicate via HTTP REST. Now, we're going to extend our demo to support gRPC communication between services and explore Linkerd's gRPC-specific features.

## What We've Built So Far

Let's review our demo application architecture:

1. **foo** - An entry point service that makes parallel requests to bar and baz
2. **bar** - A middle-tier service that processes requests and calls baz
3. **baz** - A backend service that returns request information

These services are all written in Go and deployed to a local Kubernetes cluster. We're using Tilt for a fast inner development loop, which allows us to see our changes in real-time as we modify the code.

Our architecture currently looks like this:

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│         │     │         │     │         │
│  nginx  │────►│   foo   │────►│   bar   │
│         │     │         │     │         │
└─────────┘     └────┬────┘     └────┬────┘
                     │               │
                     │               │
                     │               ▼
                     │          ┌─────────┐
                     │          │         │
                     └─────────►│   baz   │
                                │         │
                                └─────────┘
```

_Note: when running locally there is also a `synthetic` service that is constantly driving high traffic through the `foo` service to simulate application load._

## The gRPC Implementation

Since the initial blog post (link?) we've now enhanced our application to support both HTTP REST and gRPC communication. Key components of our implementation include:

### 1. Protocol Buffer Definitions

We created a shared Protobuf schema that defines our service interfaces:

```protobuf
// services.proto
syntax = "proto3";

package demo;
option go_package = "./proto";

service Foo {
  rpc GetInfo (InfoRequest) returns (InfoResponse);
}

service Bar {
  rpc GetInfo (InfoRequest) returns (InfoResponse);
}

service Baz {
  rpc GetInfo (InfoRequest) returns (InfoResponse);
}

message InfoRequest {
  string client = 1;
  map<string, string> headers = 2;
}

message InfoResponse {
  string message = 1;
  string hostname = 2;
  map<string, string> headers = 3;
  int32 status = 4;
}
```

### 2. Dual Protocol Support

Each service now supports both HTTP REST and gRPC:

- Each service runs an HTTP server on port 8080
- Each service runs a gRPC server on port 9090
- A `USE_GRPC` environment variable controls which protocol is used for service-to-service communication
- Kubernetes configuration has been updated to support the dual protocol of our services

### 3. Tilt Configuration

We've enhanced our Tilt setup to make switching between protocols and service mesh options easy:

```sh
# Run with HTTP REST (default)
tilt up

# Run with gRPC (no Linkerd)
tilt up -- --use_grpc

# Run with HTTP REST and Linkerd
tilt up -- --use_linkerd

# Run with gRPC and Linkerd
tilt up -- --use_grpc --use_linkerd
```

These configuration flags allow us to test our application in different modes, which is especially useful for comparing performance and behavior across different configurations.

## Leveraging Linkerd with gRPC

A simple demonstration of the power of linkerd is to test the behavior when running gRPC without Linkerd `tilt up -- --use_grpc`. When you run the application you'll notice that the `foo` service will only interact with a single instance of the `baz` service. This is because by default gRPC requests will not be load balanced across services. See [gRPC Load Balancing on Kubernetes without Tears](https://linkerd.io/2018/11/14/grpc-load-balancing-on-kubernetes-without-tears/) for a deep dive on why this is the case. Suffice to say we need to enable Linkerd so that we can effectively utilize gRPC without losing our load balancing features. So re-run the app with `--use_linkerd` as well.

```sh
tilt up -- --use_grpc --use_linkerd
```

### A Virtual Tour

With our application running with both grpc and linkerd enabled we can start to explore some of the benefits of Linkerd.

From the graphic below we can see two things. First, we're still getting proper request and TCP metrics for our application without the need to re-implement monitoring. This is managed by Linkerd automatically. Second, we can verify that we're now getting proper load balancing across all of our baz pods. Without Linkerd in the picture requests would have been pinned to a single instance which is not ideal, especially when working with low numbers of pods where the law of averages doesn't kick in.

![](grpc_loadbalancing.png)

From the next graphic we can also see live calls coming to our service. This is particularly useful when dealing with gRPC as we may have obfuscated our visibility into service behavior during our transition to the new protocol.

![](grpc_livecalls.png)

Don't forget, while the visuals of the dashboard are nice, the Linkerd CLI provides access to the same data.

```bash
$ linkerd viz top deployment/baz
Source                Destination           Method      Path                Count    Best   Worst    Last  Success Rate
foo-64798767b7-x8xvf  baz-659dbf6895-v7gdm  POST        /demo.Baz/GetInfo    1187    81µs     9ms   124µs       100.00%
bar-577c4bf849-cpdxl  baz-659dbf6895-9twg9  POST        /demo.Baz/GetInfo    1103    86µs     6ms   140µs       100.00%
foo-64798767b7-x8xvf  baz-659dbf6895-7chgx  POST        /demo.Baz/GetInfo    1084    79µs     9ms   355µs       100.00%
bar-577c4bf849-cpdxl  baz-659dbf6895-7chgx  POST        /demo.Baz/GetInfo    1061    93µs     7ms   293µs       100.00%
foo-64798767b7-x8xvf  baz-659dbf6895-9twg9  POST        /demo.Baz/GetInfo    1042    81µs     5ms   565µs       100.00%
bar-577c4bf849-cpdxl  baz-659dbf6895-v7gdm  POST        /demo.Baz/GetInfo     958    75µs     8ms   203µs       100.00%
```

### gRPC Authorization

Next let's implement an authorization policy for our new gRPC routes on the `baz` service. The authorization policy will give us fine grained controls for which services have permissions to access our service. But first we need to implement a new server and route. The server is going to represent the new gRPC endpoint that we've opened on the baz service. Like the HTTP route, we'll go with a default deny policy so we can verify that our services have proper authrization.

```yaml
---
apiVersion: policy.linkerd.io/v1beta1
kind: Server
metadata:
  name: baz-grpc
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: baz
  port: grpc # target the toxic server (defined in the service)
  proxyProtocol: gRPC
  accessPolicy: deny # deny all traffic by default
```

Once we have our server in place, we will follow the pattern we used in our HTTP implementation and give the `foo` server full access to our new gRPC server.

```yaml
---
apiVersion: policy.linkerd.io/v1beta1
kind: ServerAuthorization
metadata:
  namespace: default
  name: baz-grpc
spec:
  server:
    name: baz-grpc
  client:
    meshTLS:
      identities:
        - "foo.default.serviceaccount.identity.linkerd.cluster.local"
```

```yaml
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GRPCRoute
metadata:
  name: baz-grpc-inbound
  namespace: default
spec:
  parentRefs:
    - name: baz-grpc
      kind: Server
      group: policy.linkerd.io
  rules:
    - matches:
        - path:
            type: Exact
            value: "/demo.Baz/GetInfo"
```
