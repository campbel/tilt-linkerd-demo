# tilt-linkerd-demo

A demo of using Tilt to run an application with Linkerd on your local machine

## Quick Setup

1. Setup local Kubernetes cluster

   Recommend using [Orbstack](https://orbstack.dev/) (macOS only) for local Kubernetes

   Other options:

   - [Docker Desktop](https://www.docker.com/products/docker-desktop/) (macOS / Windows)
   - [Rancher Desktop](https://rancherdesktop.io/) (macOS / Windows)

   Once Kubernetes is running, verify it's working and you can connect to it with `kubectl`

   ```sh
   kubectl get nodes
   ```

2. Install Tilt

   ```sh
   brew install tilt-dev/tap/tilt
   ```

   _Full instructions are available [here](https://docs.tilt.dev/install.html)_

3. Setup `/etc/hosts` (optional, choose one of the following)

   (option A) Add the following to `/etc/hosts`:

   ```sh
   127.0.0.1   linkerd.localhost
   127.0.0.1   foo.localhost
   ```

   (option B) use `hostctl`

   ```sh
   brew install guumaster/tap/hostctl
   sudo hostctl add tilt-linkerd-demo < .etchosts
   ```

4. Install Linkerd cli (optional)

   ```sh
   curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install-edge | sh
   export PATH="$PATH:/Users/username/.linkerd/bin"
   ```

   _Full instructions are available [here](https://linkerd.io/2.16/getting-started/#step-1-install-the-cli)_

5. Install Protocol Buffer tools (for gRPC)

   ```sh
   brew install protobuf
   brew install bufbuild/buf/buf
   brew install protoc-gen-go protoc-gen-go-grpc
   ```

## Run the demo

```sh
tilt up
```

## Configuration Options

### gRPC Communication

You can enable gRPC communication between services:

#### Using the Tilt Command Line Flag

```sh
tilt up -- --use_grpc
```

This sets `USE_GRPC=true` for all services.

#### Using the Tilt Web UI

After starting Tilt, click the "Settings" icon (gear) in the top right of the Tilt web UI, then toggle the "use_grpc" setting.

#### Manually Editing Deployments

```sh
# Edit specific deployments:
kubectl edit deployment foo
```

Find the environment variables section and change `USE_GRPC` to `true`.

### Linkerd Service Mesh

By default, Linkerd is disabled. You can enable it when needed:

#### Using the Tilt Command Line Flag

```sh
# Enable Linkerd
tilt up -- --use_linkerd

# Enable both Linkerd and gRPC
tilt up -- --use_linkerd --use_grpc
```

#### Using the Tilt Web UI

After starting Tilt, click the "Settings" icon (gear) in the top right of the Tilt web UI, then toggle the "use_linkerd" setting.

### All Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `--use_grpc` | `false` | Enable gRPC communication between services |
| `--use_linkerd` | `false` | Enable Linkerd service mesh features |