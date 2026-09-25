```mermaid
stateDiagram-v2
    state "inc" as inc
    state "double" as double
    state "start" as start
    state "end" as end
    start --> inc : 
    inc --> double : 
    double --> end : 

```