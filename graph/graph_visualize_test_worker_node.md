```mermaid
stateDiagram-v2
    state "end" as end
    state "worker" as worker
    state "start" as start
    state router_start <<choice>>
    worker --> end : 
    start --> router_start : 
    router_start --> worker : 

```