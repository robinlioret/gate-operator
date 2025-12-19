# Get started

## Installation

### Prerequisite

You'll need to have a valid installation of [Cert-Manager](https://cert-manager.io/docs/installation/).

### Using the YAML installer

Run the following command: 

```shell
kubectl apply -f https://github.com/robinlioret/gate-operator/releases/download/v0.0.6/install.yaml
```

### Using the Helm chart

```shell
helm upgrade --install gate-operator oci://ghcr.io/robinlioret/gate-operator/gate-operator:0.0.5
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
    - name: KubeProxy
      selector:
        apiVersion: apps/v1
        kind: DaemonSet
        name: kube-proxy
        namespace: kube-system
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
    status: "False"
    type: Closed
  
  # Next time the gate will be evaluated
  nextEvaluation: "2025-10-31T08:51:32Z"
  
  # Quick information of the gate status (Opened or Closed)
  state: Opened
  
  # Information on each target specified. Can help for troubleshooting.
  # The condition's type field matches the name field on each target.
  targetConditions:
  - lastTransitionTime: "2025-10-31T08:50:32Z"
    message: |
      1 object(s) found
      1 object match the validators
      1/1 valid objects
    reason: ObjectsFound
    status: "True"
    type: KubeProxy
```

## ArgoCD configuration

You need to add the following snippet to your argocd cm

```yaml
resource.customizations:
  "gate.sh/*":
      health.lua: |
        hs = {}
        if obj.status ~= nil then
          if obj.status.conditions ~= nil then
            for i, condition in ipairs(obj.status.conditions) do
              if condition.type == "Opened" and condition.status == "True" then
                hs.status = "Healthy"
                hs.message = condition.message
                return hs
              end
            end
          end
        end
        hs.status = "Progressing"
        hs.message = "Waiting for the gate to open"
        return hs
```

This customizes the healthcheck for the gates, allowing ArgoCD to consider gates Healthy if opened or Progressing if closed.