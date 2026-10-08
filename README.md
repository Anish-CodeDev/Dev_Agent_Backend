# Dev_Agent Sandbox Server 🛡️

A Go gRPC backend for [Dev_Agent](https://github.com/Anish-CodeDev/Dev_Agent). The agent generates code and project files; this server receives them, persists them to a mounted volume, runs commands against them, and lets the agent view individual files it chooses. Everything runs in a [gVisor](https://gvisor.dev/)-sandboxed pod on a local [kind](https://kind.sigs.k8s.io/) cluster, so AI-generated code never touches your host directly.

---

## Features

- 📁 **File creation**: receives generated files from the agent and writes them to a persistent volume mount.
- ⚙️ **Command execution**: runs commands inside the sandbox and returns the output.
- 👀 **File viewing**: lets the agent read back individual files from the workspace, as decided by the agent.
- 🔒 **gVisor sandboxing**: the server pod runs under the `runsc` runtime, which places a user-space kernel between untrusted code and the host.
- ☸️ **Kind-based deployment**: a reproducible, local Kubernetes setup.

---

## Architecture

```
┌──────────────────┐        gRPC         ┌───────────────────────────────────────┐
│   Dev_Agent      │ ──────────────────► │  kind cluster (gVisor-enabled nodes)  │
│   (Python)       │                     │  ┌─────────────────────────────────┐  │
│                  │ ◄────────────────── │  │ Pod (runtimeClassName: gvisor)  │  │
└──────────────────┘       results       │  │  Go gRPC server                 │  │
                                         │  │  └─ workspace ◄── PVC mount     │  │
                                         │  └─────────────────────────────────┘  │
                                         └───────────────────────────────────────┘
```

1. The agent generates a file and calls `CreateFiles`.
2. The server saves it to the workspace volume.
3. The agent calls `ExecuteCommands` to install dependencies, run tests, and so on.
4. The agent calls `ViewFile` to read any individual file it wants to inspect.

---

## gRPC API

| RPC | Description |
| --- | --- |
| `CreateFiles` | Saves one or more generated files to the workspace volume |
| `ExecuteCommands` | Runs commands inside the sandboxed pod |
| `ViewFile` | Returns the contents of an individual file, chosen by the agent |

---

## Repository Layout

| Path | Purpose |
| --- | --- |
| `backend/` | The Go gRPC server: the gRPC handler and the definitions of the gRPC service functions (`CreateFiles`, `ExecuteCommands`, `ViewFile`), plus the Dockerfiles used to build the images |
| `kind_config.yaml` | kind cluster definition using the gVisor-enabled node image |
| `runtime.yaml` | `RuntimeClass` that maps `gvisor` to the `runsc` handler |
| `pvc.yaml` | PersistentVolumeClaim for the workspace volume |
| `server-deployment.yaml` | Deployment of the gRPC server with `runtimeClassName: gvisor` |

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- Go (to build the server)

---

## Setup

### Why this differs from a normal cluster

kind nodes are Docker containers with containerd running inside them. gVisor (`runsc`) must be installed *inside* the node container, and containerd needs a runtime handler registered for it. There is no separate VM boundary, so you are nesting sandboxing inside kind's own container.

### 1. Build a kind node image with gVisor baked in

The `Dockerfile.gvisor-node` in `backend/` builds a node image on top of `kindest/node` with gVisor (`runsc`) installed. Build it:

```bash
cd backend
docker build -t kind-node-gvisor:latest -f Dockerfile.gvisor-node .
```

> Pin the `kindest/node` tag to match your kind CLI version. Mismatches cause hard-to-debug cluster bring-up failures.

### 2. Create the kind cluster

`kind_config.yaml` points the nodes at the `kind-node-gvisor:latest` image and patches containerd at cluster-creation time to register the `runsc` runtime, so nothing needs hand-editing after boot.

```bash
kind create cluster --name my-cluster --config kind_config.yaml
```

Once the cluster is up, complete [Configuring `runsc.toml`](#configuring-runsctoml) before continuing. It is required.

### 3. Build and load the server image

```bash
cd backend
docker build -t dev-agent:latest -f Dockerfile.grpc .
kind load docker-image grpc-server:latest --name my-cluster
```

> The image is built as `dev-agent:latest` but loaded as `grpc-server:latest`. Make sure the tag you build, the tag you load, and the `image:` in `server-deployment.yaml` all match, otherwise the pod will end up in `ImagePullBackOff`.

### 4. Apply the manifests

Apply them in this order:

```bash
kubectl apply -f runtime.yaml            # RuntimeClass: gvisor -> runsc
kubectl apply -f pvc.yaml                # workspace volume
kubectl apply -f server-deployment.yaml  # gRPC server (runtimeClassName: gvisor)
```

Check that the pod comes up:

```bash
kubectl get pods -o wide
```

### 5. Verify the sandbox

```bash
kubectl exec -it <pod> -- dmesg | head -1
# Should show a gVisor-specific fake kernel banner, not your real host kernel
```

---

## Configuring `runsc.toml`

**This step is required.** The containerd patch in `kind_config.yaml` points the `runsc` runtime at `/etc/containerd/runsc.toml`, so the file must exist on every node before you deploy anything. Run these steps after creating the cluster (and any time you change the config), repeating for each node. `systrap` is generally the fastest platform for trapping syscalls (faster than `ptrace`); fall back to `ptrace` only if you hit compatibility issues.

1. Identify the node containers (with `my-cluster` they are `my-cluster-control-plane` and `my-cluster-worker`). If a pod is already running, you can see which node it landed on in the `NODE` column:

   ```bash
   kubectl get nodes
   kubectl get pod <pod-name> -o wide
   ```

2. Check whether the config file already exists on that node:

   ```bash
   docker exec -it <NODE> cat /etc/containerd/runsc.toml
   ```

3. If it's missing or out of date, write it and restart containerd:

   ```bash
   docker exec -it <NODE> bash -c 'cat <<EOF > /etc/containerd/runsc.toml
   [runsc_config]
     platform = "systrap"
     network = "host"
   EOF'
   docker exec -it <NODE> systemctl restart containerd
   ```

---

## Connecting from Dev_Agent

The gRPC server is available at **`localhost:9000`**. Point the agent's gRPC client at that address and call `CreateFiles`, `ExecuteCommands` and `ViewFile`.

If the port isn't already exposed through your cluster setup, forward it:

```bash
kubectl port-forward deploy/dev-agent-deployment 9000:9000
```

---

## Debugging

Enable gVisor debug logs by adding this to `runsc.toml`:

```toml
[runsc_config]
  debug = true
  debug-log = "/tmp/runsc/%ID%/"
```

Then inspect the logs on the node:

```bash
docker exec -it my-cluster-worker bash
ls /tmp/runsc/
```

You can also use `crictl inspect <container-id>` inside the node to confirm the runtime handler was applied.

---

## Common Pitfalls

- **Unpinned node image:** forgetting to match the `kindest/node` tag to your kind CLI version breaks cluster bring-up.
- **Missing `runtimeClassName`:** applying the RuntimeClass but not setting `runtimeClassName` on the pod spec means the pod silently runs unsandboxed.
- **No latency testing:** netstack overhead is the thing most likely to surprise you. Load-test gRPC latency before and after enabling gVisor.

---

## Security Notes

- gVisor significantly reduces host attack surface but isn't a silver bullet. Keep the cluster local and treat it as a development tool.
- The server should resolve incoming file paths against the workspace root and reject anything that escapes it (`../`, absolute paths, symlinks).
- If you expose the gRPC port beyond localhost, add TLS and authentication.

---

## Related

- 🤖 [Dev_Agent](https://github.com/Anish-CodeDev/Dev_Agent): the multi-agent framework that uses this server

## Contributing

PRs and issues welcome.