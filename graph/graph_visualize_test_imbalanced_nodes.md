```mermaid
stateDiagram-v2
    state "start" as start
    state "end" as end
    state "inc" as inc
    state "inc2" as inc2
    state "double" as double
    inc2 --> double : 
    inc --> end : 
    double --> end : 
    start --> inc : 
    start --> inc2 : 

```