# Get started

## Installation

### Using the YAML installer

Run the following command: 

```shell
kubectl apply -f https://raw.githubusercontent.com/robinlioret/gate-operator/refs/heads/main/dist/install.yaml
```

### Using the Helm chart

```shell
helm upgrade --install ghcr.io/robinlioret/gate-operator/gate-operator:0.0.1 gate-operator
```

Possible values can be found here: https://github.com/robinlioret/gate-operator/blob/main/dist/chart/values.yaml

## Create your first gate

```shell
cat << EOF | kubectl apply -f -
apiVersion: gate.sh/v1alpha1
kind: Gate
metadata:
  name: gate-test
spec:
  targets:
    - targetName: KubeProxy
      apiVersion: apps/v1
      kind: DaemonSet
      name: kube-proxy
      namespace: kube-system
      existsOnly: true
EOF
```

This gate should open if kube-proxy is deployed in the namespace kube-system.

```shell
kubectl get gates -A
```

Should display :
```
NAMESPACE   NAME        STATE
default     gate-test   Opened
```

## Reading information on your gates

You can get information on you gate with the following command

```shell
kubectl get gate gate-test -o yaml
```

```yaml
apiVersion: gate.sh/v1alpha1
kind: Gate
metadata:
  name: gate-test
  namespace: default
  # ...
spec:
  # ...
status:
  # Condition of the gate for programmatic access
  conditions:
  - lastTransitionTime: "2025-10-31T08:40:26Z"
    message: Gate was evaluated to true
    reason: GateConditionMet
    status: "True"
    type: Opened
  - lastTransitionTime: "2025-10-31T08:40:26Z"
    message: Gate was evaluated to true
    reason: GateConditionMet
    status: "True"
    type: Available
  - lastTransitionTime: "2025-10-31T08:40:26Z"
    message: Gate was evaluated to true
    reason: GateConditionMet
    status: "False"
    type: Closed
  - lastTransitionTime: "2025-10-31T08:40:26Z"
    message: Gate was evaluated to true
    reason: GateConditionMet
    status: "False"
    type: Progressing
  
  # Next time the gate will be evaluated
  nextEvaluation: "2025-10-31T08:51:32Z"
  
  # Quick information of the gate status (Opened or Closed)
  state: Opened
  
  # Information on each target specified. Can help for troubleshooting.
  # The condition's type field matches the targetName field on each target.
  targetConditions:
  - lastTransitionTime: "2025-10-31T08:50:32Z"
    message: 1 object(s) found
    reason: ObjectsFound
    status: "True"
    type: KubeProxy
```

