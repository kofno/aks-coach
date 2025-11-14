# aks-coach

Tiny CLI to snapshot **Kubernetes capacity** by Deployment + **HPA min/max/targets**.

## Install
Download a binary from the [Releases], or build from source:
```bash
go install github.com/yourname/aks-coach/cmd/aks-coach@latest
```

## Usage

```bash
# Single namespace
aks-coach -n demo

# All namespaces
aks-coach --all-namespaces

# Filter by label
aks-coach -n production -l app=mywebapp

# JSON output (pipe to jq)
aks-coach --all-namespaces -o json | jq .

```

Shows per-Deployment:
- replicas, CPU/mem requests & limits (summed × replicas)
- HPA min/max and CPU target like cpu: 23%/60%
- (Title shows scope + selector)

## Completions

```bash
aks-coach completion bash > /usr/local/etc/bash_completion.d/aks-coach
aks-coach completion zsh  > /usr/local/share/zsh/site-functions/_aks-coach
```

## License
MIT

