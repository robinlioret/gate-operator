# Reference

The Gate and ClusterGate CRD share the exact same reference and are interchangeable.

```yaml
apiVersion: gate.sh/v1alpha1
kind: Gate # or ClusterGate
metadata:
  # ...
spec:
  # Targets are rules to fulfill to open the gate.
  # A gate must have at least one target.
  targets:
      # (Optional) Default to Target{index}
      # Name of the target, used to define target condition Type field
    - name: ATargetName 
      # (Required) Rules used to find resource to evaluate
      selector:
        # (Required) Api Version of the resource
        apiVersion: apps/v1
        # (Required) Kind of object to look for
        kind: Deployment
        # (Optional) Namespace of the object.
        # By default, and if it's relevant, the gate looks for resources inside its own namespace
        namespace: my-namespace
        # (Optional and mutually exclusive with labelSelector)
        # Name of the resource.
        name: my-deployment
        # (Optional and mutually exclusive with name)
        # Specification of a label selector
        labelSelector:
          matchLabels: {}
          matchExpression: {}
      # (Optional) Instruction on how to evaluate the target
      # By default, the gate will open if at least one resource was found regardless of its state
      # Validators are "anded" together (see §Behavious and patterns of validators)
      validators:
        # (Optional) Minimal amount of valid objects to validate the target. By default, it a minimum of 1.
        - atLeast:
            # (Optional) At least N objects must validate all the given conditions
            count: 1
            # (Optional) At least X percent of the found object must validate the other validators (if there are ones)
            percent: 50
        # (Optional) Check if the object has the right condition in its status.conditions
        - matchCondition:
            # (Required) Condition type in CamelCase
            # Literally the type field of the condition to look for
            type: Ready
            # (Optional) default to "True" (with quotes, it's a string)
            # The value of the condition "True", or "False"
            status: "True"
        # (Optional) Check if the field designed by the json pointer matches the given value
        - jsonPointer:
            # (Required) The JSON pointer (see https://gregsdennis.github.io/Manatee.Json/usage/pointer.html)
            pointer: /status/state
            # (Required) Value the JSON pointer must have to validate the validator
            value: Opened
  # (Optional) Operation to perform to reduce the targets to a single boolean
  # By default, the targets are "anded"
  operation:
    # (Required) The operation's name
    # Must be And or Or
    operator: And
  # (Optional) Default to 1 minutes (1m0s)
  # Delay between two evaluations
  # WARNING: subject to change
  requeueAfter: 53s
# (Managed) status field with the computed resources on the gate
status:
  # Quick representation of the gate's status
  # Used by the print feature
  state: Opened # or Closed
  # Kubernetes' conditions of the gate
  conditions: # ...
  # Result of the target computation
  # Here can be found useful information for troubleshoot purposes
  targetConditions:
    - type: ATargetName # From the name of the represented target
      status: "False"   # Condition status
      reason: "ConditionNotMet"
      message: |
        [my-namespace/my-deployment] condition Ready is not True
        1 object(s) found
        0 object(s) validated
        0/1 valid object(s)
      # ...
```

## Behaviour and patterns of validators

There are three scenarios regarding the atLeast validator.

1. One validator and it is atLeast

The gate will open if at least N objects were selected regardless of their state

2. One or more validator(s) but no atLeast validator

The gate will open if at least 1 object was found and all the objects validate the validators.

Per example: all the pod matching the label "app" = "my-app" has a Ready condition to true. The minimum at 1 protects against "false opening" when resources are not yet deployed.

3. Two or more validator(s) with atLeast

The gate will open if atLeast N (or X%) objects fulfill the other validators.