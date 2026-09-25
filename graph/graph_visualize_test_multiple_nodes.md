```mermaid
stateDiagram-v2
    state "end" as end
    state "inc" as inc
    state "double" as double
    state "start" as start
    double --> end : 
    start --> inc : 
    inc --> double : 

```