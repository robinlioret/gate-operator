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
    - targetName: MyDeployment
      apiVersion: apps/v1
      kind: Deployment
      name: my-deployment
      desiredCondition:
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
    - targetName: Gates
      apiVersion: gate.sh/v1alpha1
      kind: Gate
      labelSelector:
        matchLabels:
          deployment-stage: stage-x
      desiredCondition:
        type: Opened
```