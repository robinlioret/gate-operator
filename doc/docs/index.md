# Gate Operator

The Gate Operator solves a simple problem : orchestration among declarative kubernetes workloads.

## Why

Even though eventual consistency is a core principle of Kubernetes philosophy, it is not always achievable in a purely declarative way.

- One pod may start with default credentials before the said credentials could be retrieved. Leading to a soft lock of the pod.
- Another workload may create a LoadBalancer service before the load balancer controller was deployed. Leading to the usage of legacy controller and misconfiguration (Hello, EKS!)

Why is eventual consistency not always achievable is open to discussion. But, it is here and we need to deal with it.

## Solution

To solve this issue, tools exists. ArgoCd, per example, provide SyncWaves and Hooks. However, it is only applicable on the application scope or, more recently, on the ApplicationSet scope. It's a partial solution. 
In the same way, Helm provides char hooks. But, again, it's only usable in the chart scope.

These limitations make difficult to orchestrate larger workloads spread across multiple charts, applications, etc.

Enters Gate Operator with a simple yet powerful concept to synergize with deployment tools (Helm, etc)

## Concept

A Gate is a resource that have two possible state: opened or closed.

It changes state when its logicial expression is validated (then become opened) or invalidated (the become closed).

It has a logicial expression about resources in the cluster being in a certain way.

Per exemple:

- A gate may wait for a specific Deployment to be Available
- Or to wait for a ConfigMap to be created
- Or for another gate to be Opened (or Closed)

```mermaid
flowchart LR
    classDef wave1 fill:#FDE7FE,stroke:#333,stroke-width:3px;
    classDef wave2 fill:#FDE7FE,stroke:#333,stroke-width:3px;
    classDef wave3 fill:#F9BBFC,stroke:#333,stroke-width:3px;
    classDef wave4 fill:#F48FFA,stroke:#333,stroke-width:3px;
    classDef wave5 fill:#F063F8,stroke:#333,stroke-width:3px;
    
    subgraph deploy-a [Deployment A]
        direction LR
        deploy-a-res1[Deployment]
        deploy-a-res2[Secret]
        deploy-a-res3[Resources...]
        deploy-a-gate-out(("Gate"))
        deploy-a-gate-out --> deploy-a-res1 & deploy-a-res2 & deploy-a-res3
        
        class deploy-a-res1 wave1;
        class deploy-a-res2 wave1;
        class deploy-a-res3 wave1;
        class deploy-a-gate-out wave1;
    end
    
    subgraph deploy-b [Deployment B]
        direction LR
        deploy-b-res1[Deployment]
        deploy-b-res2[Secret]
        deploy-b-res3[Resources...]
        deploy-b-gate-out(("Gate"))
        deploy-b-gate-out --> deploy-b-res1 & deploy-b-res2 & deploy-b-res3

        class deploy-b-res1 wave2;
        class deploy-b-res2 wave2;
        class deploy-b-res3 wave2;
        class deploy-b-gate-out wave2;
    end
    
    subgraph deploy-c [Deployment C]
        direction LR
        deploy-c-gate-in(("Gate"))
        deploy-c-res1[Deployment 1]
        deploy-c-res2[Deployment 2]
        deploy-c-gate-internal(("Gate"))
        deploy-c-res3[Deployment 3]
        deploy-c-res4[Standalone<br/>Deployment]
        deploy-c-gate-out(("Gate"))
        deploy-c-gate-in --> deploy-b-gate-out
        deploy-c-gate-in ----> deploy-a-gate-out
        deploy-c-res1 & deploy-c-res2 --> deploy-c-gate-in
        deploy-c-gate-internal --> deploy-c-res1 & deploy-c-res2
        deploy-c-res3 --> deploy-c-gate-internal
        deploy-c-gate-out -----> deploy-c-res4
        deploy-c-gate-out --> deploy-c-res3
        
        class deploy-c-gate-in wave3;
        class deploy-c-res4 wave3;
        
        class deploy-c-res1 wave4;
        class deploy-c-res2 wave4;
        class deploy-c-gate-internal wave4;

        class deploy-c-res3 wave5;
        class deploy-c-gate-out wave5;
    end
    
    other["Other"] --> deploy-c-gate-out
    
    class other wave5;
```

## Synergy

- Helm: a gate can be integrated in an application chart with the annotation `helm.sh/hook: pre-install` to prevent the creation of the chart resource before other resources are up and running.
- ArgoCD: same thing with Hook or SyncWaves. SyncWaves allows a gate to "pause" the deployment in the middle.