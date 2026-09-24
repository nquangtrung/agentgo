```mermaid
stateDiagram-v2
    state "triple" as triple
    state "start" as start
    state "end" as end
    state "inc" as inc
    state "inc2" as inc2
    state "double" as double
    triple --> end : 
    inc2 --> end : 
    start --> inc : 
    inc --> inc2 : 
    inc --> double : 
    inc --> triple : 
    double --> end : 

```