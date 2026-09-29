```mermaid
stateDiagram-v2
    state "start" as start
    state "end" as end
    state "inc" as inc
    state "double" as double
    inc --> double : 
    double --> end : 
    start --> inc : 

```