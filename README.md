# Image Pull Secret Injector

A Kubernetes mutating webhook that automatically injects image pull secrets into pods that need to pull images from Docker Hub. This helps manage Docker Hub's rate limiting by ensuring all pulls are authenticated.

## Overview

The Image Pull Secret Injector is a mutating webhook that:
- Automatically injects Docker Hub pull secrets into pods
- Synchronizes the pull secret to the pod's namespace
- Provides flexibility to exclude specific pods from secret injection
- Helps organizations manage Docker Hub's rate limits effectively

## Features

### 1. Automatic Secret Injection

The webhook automatically injects image pull secrets into pods that:
- Reference Docker Hub images
- Don't have pull secrets already specified

Example of a pod that will receive the pull secret:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: nginx
    image: nginx:latest
```

After the webhook processes this pod, it will automatically:
1. Create a copy of the configured pull secret in the pod's namespace
2. Inject the pull secret reference into the pod specification

### 2. Pod Exclusion

You can exclude specific pods from secret injection using an annotation.

Example of an excluded pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: excluded-pod
  annotations:
    image-pull-secret-injector.facets.cloud/exclude: "true"
spec:
  containers:
  - name: nginx
    image: nginx:latest
```

The annotation `image-pull-secret-injector.facets.cloud/exclude: "true"` tells the webhook to skip this pod during secret injection.

### 3. Secret Synchronization

The webhook automatically:
- Copies the source pull secret to the target namespace
- Maintains the secret's content in sync with the source
- Handles secret updates and deletions

## Requirements

For deployment, ensure you have:
- A recent version of kustomize in your PATH
- cert-manager installed in your cluster (for automatic webhook TLS setup)

## Installation

### Using Make

```bash
# Deploy the webhook
make deploy

# Remove the webhook
make undeploy
```

### Manual Installation

1. Clone the repository
2. Configure your Docker Hub credentials in a secret
3. Apply the webhook configuration
4. Deploy the webhook server

## Configuration

### Secret Configuration

The webhook can be configured to inject multiple secrets through the `--secrets` command-line argument in the deployment configuration. Edit the `config/webhook/deployment.yaml` file to specify your secrets:

```yaml
spec:
  template:
    spec:
      containers:
      - name: mutator
        args:
        - --secrets=secret1,secret2,secret3  # List your secrets here
        - --debug=true
```

Example:
```yaml
args:
- --secrets=registry-secret-dockerhub,registry-secret-gcr,registry-secret-ecr
- --debug=true
```

Each secret in the list will be:
1. Copied from its source namespace to the target pod's namespace
2. Added to the pod's `imagePullSecrets` list

Other key configurations include:
- Source secret name and namespace
- Target secret name
- Excluded namespaces
- TLS configuration

## Troubleshooting

Common issues and solutions:

1. **Pods not getting secrets:**
   - Verify the webhook is running
   - Check pod annotations for exclusions
   - Ensure the source secret exists

2. **TLS Issues:**
   - Verify cert-manager is installed
   - Check webhook service and certificate resources

3. **Rate Limiting Still Occurs:**
   - Confirm Docker Hub credentials are valid
   - Verify secret synchronization
   - Check pod events for injection status

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Submit a pull request

## License

This project is licensed under the terms specified in the repository.
