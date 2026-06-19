# Key Components of Harness: A Three-Layer Architecture

## ASCII Diagram

```text
+---------------------------------------------------------------------------------------+
|                                    Information Layer                                  |
|   Determine what an agent starting work, what information it can see, and what        |
|                            capabilities it can invoke                                 |
|                                                                                       |
|  +---------------------------------------+         +-------------------------------+  |
|  |     Memory & Context Management       | ------> |        Tools & Skills         |  |
|  | Long term/Short term Memory, Knowledge|         | Available APIs, external      |  |
|  | Base, Conversation Context            |         | tools, and specialized skills |  |
|  +---------------------------------------+         +-------------------------------+  |
+---------------------------------------------------------------------------------------+
                                        |
                                        v
+---------------------------------------------------------------------------------------+
|                                     Execution Layer                                   |
|   Determine how work is decomposed, how agents collaborates, where the boundaries     |
|                   lie, how to recover in case of failure, etc                         |
|                                                                                       |
|  +---------------------------------------+         +-------------------------------+  |
|  |     Orchestration & Coordination      | <-----> |       Infra & Guardrails      |  |
|  | work flow management, task decomposi- |         | Operating environment, resou- |  |
|  | tion, agent collaboration, planning   |         | rce limitations, security...  |  |
|  +---------------------------------------+         +-------------------------------+  |
+---------------------------------------------------------------------------------------+
                                        |
                                        v
+---------------------------------------------------------------------------------------+
|                                     Feedback Layer                                    |
|   Determines how well the system can improve over time, whether the results of each   |
|     execution are verified, whether each failure is recorded and transformed, etc     |
|                                                                                       |
|  +---------------------------------------+         +-------------------------------+  |
|  |      Evaluation & Verification        |         |     Tracing & Observability   |  |
|  | Result verification, quality assess-  |         | Logging, performance monitor- |  |
|  | ment, and correctness check           |         | ing, debugging info...        |  |
|  +---------------------------------------+         +-------------------------------+  |
+---------------------------------------------------------------------------------------+
```

## Mermaid Diagram

```mermaid
graph TD
    subgraph Information_Layer [Information Layer]
        IL_Desc["Determine what an agent starting work, what information it can see, and what capabilities it can invoke"]
        Memory["Memory & Context Management<br/>Long term/Short term Memory, Knowledge Base, Conversation Context"]
        Tools["Tools & Skills<br/>Available APIs, external tools, and specialized skills"]
        
        Memory --> Tools
    end

    subgraph Execution_Layer [Execution Layer]
        EL_Desc["Determine how work is decomposed, how agents collaborates, where the boundaries lie, how to recover in case of failure, etc"]
        Orch["Orchestration & Coordination<br/>work flow management, task decomposition, agent collaboration, planning"]
        Infra["Infra & Guardrails<br/>Operating environment, resource limitations, security boundaries, and fault recovery"]
        
        Orch <--> Infra
    end

    subgraph Feedback_Layer [Feedback Layer]
        FL_Desc["Determines how well the system can improve over time, whether the results of each execution are verified, whether each failure is recorded and transformed, etc"]
        Eval["Evaluation & Verification<br/>Result verification, quality assessment, and correctness check"]
        Trace["Tracing & Observability<br/>Logging, performance monitoring, debugging information, failure analysis"]
    end

    Information_Layer --> Execution_Layer
    Execution_Layer --> Feedback_Layer

    style Information_Layer fill:#f9f,stroke:#333,stroke-width:2px
    style Execution_Layer fill:#bbf,stroke:#333,stroke-width:2px
    style Feedback_Layer fill:#bfb,stroke:#333,stroke-width:2px
```
