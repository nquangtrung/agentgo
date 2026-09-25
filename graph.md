```mermaid
stateDiagram-v2
    state "start" as start
    state "stream_text" as stream_text
    state "end_text" as end_text
    state "loop_check" as loop_check
    state "resolve_tool" as resolve_tool
    state "execute_tool" as execute_tool
    state "generate_text" as generate_text
    state "end_step" as end_step
    state "prepare_process" as prepare_process
    state "end_process" as end_process
    state "prepare_step" as prepare_step
    state "prepare_text" as prepare_text
    state router_resolve_tool <<choice>>
    state "end" as end
    start --> prepare_process :
    prepare_process --> prepare_step :
    loop_check --> resolve_tool : tool
    loop_check --> prepare_text : text
    loop_check --> end_step : end
    prepare_text --> generate_text : generate
    prepare_text --> stream_text : stream
    stream_text --> end_text :
    generate_text --> end_text :
    end_text --> end_step :
    prepare_step --> loop_check :
    resolve_tool --> router_resolve_tool :
    router_resolve_tool --> execute_tool :
    execute_tool --> end_step :
    end_step --> prepare_step : loop
    end_step --> end_process : end
    end_process --> end :
```
