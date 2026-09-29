```mermaid
stateDiagram-v2
    state "end" as end
    state "inc" as inc
    state "start" as start
    start --> inc : 
    inc --> end : 

```