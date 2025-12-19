# Examples

## Deployment waiter

This will open if the deployment has the Available condition set to true.

```yaml
apiVersion: gate.sh/v1alpha1
kind: Gate
metadata:
  name: wait-for-a-deployment
spec:
  targets:
    - name: MyDeployment
      selector:
        apiVersion: apps/v1
        kind: Deployment
        name: my-deployment
      validators:
        - matchCondition:
            type: Available
```

## Gate waiter

This will open if other gates with the given labels are opened. This is useful to abstract workloads away.

```yaml
apiVersion: gate.sh/v1alpha1
kind: Gate
metadata:
  name: wait-for-other-gates
spec:
  targets:
    - name: Gates
      selector:
        apiVersion: gate.sh/v1alpha1
        kind: Gate
        labelSelector:
          matchLabels:
            deployment-stage: stage-x
      validators:
        - atLeast:
            count: 3
        - matchCondition:
            type: Opened
```

## ArgoCD Application waiter

This gate will open if the argocd application `status.health.status` is set to `Healthy`

```yaml
apiVersion: gate.sh/v1alpha1
kind: Gate
metadata:
  name: wait-for-other-gates
spec:
  targets:
    - name: Application
      selector:
        apiVersion: argoproj.io/v1alpha1
        kind: Application
        name: my-app
        namespace: gitops
      validators:
        - jsonPointer:
            pointer: /status/health/status
            value: Healthy
```