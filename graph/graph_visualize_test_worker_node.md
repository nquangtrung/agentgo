```mermaid
stateDiagram-v2
    state "start" as start
    state "end" as end
    state "worker" as worker
    state router_start <<choice>>
    worker --> end : 
    start --> router_start : 
    router_start --> worker : 

```