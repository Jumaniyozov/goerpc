install kind for working with kubectl
```go install sigs.k8s.io/kind@v0.24.0```

```kind create cluster --config k8s/kind.yaml```

apply the k8s resources
```kubectl apply -f k8s/server.yaml```
```kubectl apply -f k8s/client.yaml```