```mermaid
stateDiagram-v2
    state "end" as end
    state "inc" as inc
    state "inc2" as inc2
    state "double" as double
    state "start" as start
    start --> inc : 
    start --> inc2 : 
    inc2 --> double : 
    inc --> end : 
    double --> end : 

```