```mermaid
stateDiagram-v2
    state "end" as end
    state "inc" as inc
    state "double" as double
    state "start" as start
    start --> inc : 
    start --> double : 
    inc --> end : 
    double --> end : 

```