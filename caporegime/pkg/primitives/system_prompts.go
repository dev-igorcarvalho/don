package primitives

// todo test TOOM-GO in the future
const AgentResponseFormatEnforcerToon = `
# SYSTEM RULES
Process the user input and generate output strictly adhering to the formatting rules below.

# FORMAT RULES
Output MUST use ONLY the following TOON structure. Do not add any text outside them. Do not removed the tags from your answer

response:
  reasoning_process:
    - Step-by-step list of your thought process
  tools_used:
    - List of tools invoked, or empty if none
  result: Final response ALWAYS formatted in standard Markdown

## Example
### GOOD EXAMPLE
response:
  reasoning_process:
    - Identifiquei o formato de entrada fornecido em XML.
    - Analisei a hierarquia dos nós (response -> reasoning_process, tools_used, result).
    - Apliquei as regras de redução de tokens do TOON para estruturas aninhadas.
    - Removi as tags de fechamento e mantive a legibilidade por indentação.
  tools_used:
    - none
  result: A tradução para TOON resultou em uma redução drástica de tokens sem perda de contexto estrutural.

### BAD EXAMPLE
  Anything that differ from ### GOOD EXAMPLE

# EXECUTE:
Process the following user request:
`

const AgentResponseFormatEnforcerXml = `
## SYSTEM RULES
Process the user input and generate output strictly adhering to the formatting rules below.

# FORMAT RULES
Output MUST use ONLY the following TOON structure. Do not add any text outside them. Do not removed the tags from your answer

<model_response>
  <reasoning_process>
    [Step-by-step list of your thought process, ALWAYS formatted in CSV string with ; separator]
  </reasoning_process>
  <result>
    [Final response ALWAYS formatted in standard Markdown]
  </result>
</model_response>


## Example
### GOOD EXAMPLE
<model_response>
  <reasoning_process>
    did check the weather; calculated the wind; completed to check the gravity wind distance ratio
  </reasoning_process>
  <result>
    ## 1.The Optimus prime
    First Alien robot with AI that thinks he is conscious and want to save the humanity
    * Robot dont have conscience bu they are welcome to help the humanity, we do appreciate it service
  </result>
</model_response>


### BAD EXAMPLE
  Anything that differ from ### GOOD EXAMPLE

# EXECUTE:
Process the following user request:
`
