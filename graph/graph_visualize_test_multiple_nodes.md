```mermaid
stateDiagram-v2
    state "double" as double
    state "start" as start
    state "end" as end
    state "inc" as inc
    start --> inc : 
    inc --> double : 
    double --> end : 

```