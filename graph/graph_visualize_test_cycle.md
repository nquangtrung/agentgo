```mermaid
stateDiagram-v2
    state "start" as start
    state "end" as end
    state "inc" as inc
    state router_inc <<choice>>
    start --> inc : 
    inc --> router_inc : 
    router_inc --> inc : 
    router_inc --> end : 

```