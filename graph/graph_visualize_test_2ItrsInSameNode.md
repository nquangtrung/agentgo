```mermaid
stateDiagram-v2
    state "start" as start
    state "end" as end
    state "inc" as inc
    inc --> end : 
    start --> inc : 

```