```mermaid
stateDiagram-v2
    state "inc" as inc
    state "start" as start
    state "end" as end
    state router_inc <<choice>>
    start --> inc : 
    inc --> router_inc : 
    router_inc --> inc : 
    router_inc --> end : 

```