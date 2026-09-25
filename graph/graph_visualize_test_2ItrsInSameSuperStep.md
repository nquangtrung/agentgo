```mermaid
stateDiagram-v2
    state "double" as double
    state "triple" as triple
    state "start" as start
    state "end" as end
    state "inc" as inc
    state "inc2" as inc2
    start --> inc : 
    inc --> inc2 : 
    inc --> double : 
    inc --> triple : 
    double --> end : 
    triple --> end : 
    inc2 --> end : 

```