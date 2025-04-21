# -*- mode: Python -*-
load('ext://min_tilt_version', 'min_tilt_version')
min_tilt_version('0.33.1')

load('ext://dotenv', 'dotenv')
dotenv()

# Global settings
config.define_bool("use_grpc", False, "Use gRPC for service communication")
config.define_bool("use_linkerd", False, "Enable Linkerd service mesh features")
cfg = config.parse()

# Set global environment variables
os.environ['USE_GRPC'] = "true" if cfg.get("use_grpc") else "false"
os.environ['USE_LINKERD'] = "true" if cfg.get("use_linkerd", False) else "false"

# Print feature status
if cfg.get("use_grpc"):
  print("🚀 gRPC communication ENABLED")
else:
  print("🔄 Using HTTP REST communication")

if cfg.get("use_linkerd", False):
  print("🔗 Linkerd service mesh ENABLED")
else:
  print("🔗 Linkerd service mesh DISABLED")

# Manage Contexts
context = os.environ.get('TILT_K8S_CONTEXT', 'docker-desktop')
allow_k8s_contexts(context)
current_context = k8s_context()
if current_context != context:
  warn('current k8s context is "{}" needs "{}". switching...'.format(current_context, context))
  local('kubectl config use-context {}'.format(context))


# Manage Registries
docker_registry = os.environ.get('TILT_DOCKER_REGISTRY', None)
if docker_registry:
  default_registry(docker_registry)

# Load infra resources
include('infra/ingress-nginx/Tiltfile')
include('infra/linkerd/Tiltfile')
include('infra/grafana/Tiltfile')

# Load app resources
include('apps/foo/Tiltfile')
include('apps/bar/Tiltfile')
include('apps/baz/Tiltfile')

# Load synthetic resources
include('synthetic/Tiltfile')
