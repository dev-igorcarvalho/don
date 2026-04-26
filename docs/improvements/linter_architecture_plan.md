# Plano de Implementação: Linter e Governança de Arquitetura

Este documento propõe a adoção de ferramentas de linting para garantir a qualidade do código e a integridade da arquitetura Hexagonal no projeto **Don**.

## 1. Visão Geral

Para um projeto que segue a Arquitetura Hexagonal, o linter deve ser capaz de:
1. Forçar boas práticas de Go (estilo, performance, segurança).
2. Impedir violações de dependência (ex: a camada `domain` não deve importar nada de `adapters` ou `usecases`).
3. Garantir a segregação correta entre `internal` e `pkg`.

---

## 2. Opções Analisadas

### Opção A: `golangci-lint` + `depguard` (Recomendado)
O `golangci-lint` é o agregador padrão da indústria. O linter `depguard` (embutido nele) permite definir listas brancas/pretas de imports por pacote.

*   **Prós:**
    *   Tudo em um único arquivo de configuração (`.golangci.yml`).
    *   Velocidade extrema (roda os linters em paralelo).
    *   Integração nativa com quase todos os IDEs e CI/CD.
    *   Não requer instalação de ferramentas externas além do próprio `golangci-lint`.
*   **Contras:**
    *   A configuração do `depguard` pode ficar extensa em projetos muito grandes.

### Opção B: `go-arch-lint`
Uma ferramenta dedicada exclusivamente a validar regras de arquitetura através de um arquivo YAML próprio.

*   **Prós:**
    *   Sintaxe muito expressiva e focada em "camadas".
    *   Gera gráficos de dependência.
    *   Mais fácil de ler para regras complexas de arquitetura.
*   **Contras:**
    *   Mais uma ferramenta para instalar e manter no pipeline.
    *   Não substitui o linter de código (ainda precisaria do `golangci-lint`).

---

## 3. Plano de Implementação Proposto

### Passo 1: Instalação e Configuração do `golangci-lint`
Criaremos um arquivo `.golangci.yml` na raiz do projeto com os seguintes grupos de linters:
*   **Default Linters:** `govet`, `errcheck`, `staticcheck`, `unused`.
*   **Code Quality:** `revive`, `gocritic`, `misspell`, `nolintlint`.
*   **Architecture:** `depguard`, `importas`.

### Passo 2: Definição das Regras de Camada (Hexagonal)
Configuraremos o `depguard` para forçar as seguintes restrições:

1.  **Domain:** Só pode importar `pkg` ou bibliotecas standard. NUNCA `usecases`, `adapters` ou `handlers`.
2.  **Usecases:** Pode importar `domain` e `pkg`. NUNCA `adapters` ou `handlers`.
3.  **Adapters/Handlers:** Podem importar `usecases`, `domain` e `pkg`.
4.  **Pkg:** Não deve importar nada de `internal`.

### Passo 3: Automação
Adicionar o comando de lint ao `Makefile` (ou script de CI) para falhar o build em caso de violação.

---

## 4. Como Implementar (Exemplo .golangci.yml)

Abaixo, a estrutura básica para configurar as restrições arquiteturais:

```yaml
linters-settings:
  depguard:
    rules:
      main:
        files:
          - "$all"
        deny:
          - pkg: "github.com/dev-igorcarvalho/don/internal"
            desc: "Internal packages should not be imported by pkg"
      domain:
        files:
          - "**/internal/core/domain/**"
        allow:
          - "$gostd"
          - "github.com/dev-igorcarvalho/don/pkg/**"
        deny:
          - pkg: "github.com/dev-igorcarvalho/don/internal/core/usecases"
          - pkg: "github.com/dev-igorcarvalho/don/internal/adapters"
          - pkg: "github.com/dev-igorcarvalho/don/internal/handlers"
```

---

## 5. Implementação Realizada

Os seguintes arquivos foram criados:
1.  `.golangci.yml`: Configuração completa do linter com regras de arquitetura hexagonal via `depguard`.
2.  `Makefile`: Atalhos para execução do linter e outras tarefas comuns.

### Como Usar

Para rodar o linter e verificar a arquitetura:
```bash
make lint
```

Para tentar corrigir automaticamente problemas de estilo:
```bash
make lint-fix
```

## 6. Próximos Passos (Para o Desenvolvedor)

1. [x] Aprovar o uso do `golangci-lint` com `depguard`.
2. [x] Criar o arquivo `.golangci.yml` completo.
3. [ ] Rodar `make lint` localmente e corrigir violações existentes (se houver).
4. [ ] Integrar `make lint` no pipeline de CI (GitHub Actions, GitLab CI, etc).
