# Release pelo GitHub Actions: especificação

## Problem Statement

Hoje o `SupplyCall.exe` só existe na máquina de quem roda `build.ps1`. Quem quer usar o programa precisa instalar Go e compilar. Uma action manual gera o `.exe` no GitHub, cria a tag de versão no commit exato do build e publica o `.exe` num Release público, baixável por qualquer pessoa sem login.

## Goals

- [ ] Um clique em "Run workflow" publica um Release `vX.Y.Z` com o `SupplyCall.exe` anexado.
- [ ] Todo Release aponta para o commit do qual o `.exe` foi compilado (tag + SHA nas notas).
- [ ] Qualquer pessoa baixa o `.exe` sem login, inclusive por um link fixo para a última versão.

## Out of Scope

| Feature | Reason |
| ------- | ------ |
| Release automático a cada push | Escolha do usuário: disparo manual. |
| Bump pelos Conventional Commits | Escolha do usuário: o bump é escolhido no disparo. |
| Assinatura de código (Authenticode) | Exige certificado pago. |
| Gravar o SHA do commit dentro do `.exe` | O Release e a tag já ligam o binário ao commit; mexeria em `main.go` e no menu da bandeja. |
| Builds para outras plataformas | Alvo único: Windows amd64. |
| Changelog gerado | Notas trazem versão e commit; o resto é editado à mão se preciso. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --------------------- | -------------- | --------- | ---------- |
| Gatilho | Só `workflow_dispatch`, com entrada `bump` = `patch` / `minor` / `major` (padrão `patch`) | Escolha do usuário. | y |
| Primeira versão | Sem nenhuma tag `vX.Y.Z` no repo, a versão é `v1.0.0`, qualquer que seja o bump | Escolha do usuário; o commit inicial já cita 1.0.0. | y |
| Tags consideradas | Só tags no formato exato `vMAJOR.MINOR.PATCH` (números sem zero à esquerda); o resto é ignorado | Evita que tags soltas quebrem a conta. | n |
| "Artefato do GitHub" | O `.exe` vai como asset do Release (público, sem login, não expira) **e** como artifact do run (`actions/upload-artifact`) | Artifacts de Actions exigem login para baixar e expiram; só o Release atende "qualquer pessoa". O artifact fica como pedido. | n |
| Nome do asset | `SupplyCall.exe`, sem versão no nome | Mantém fixo o link `releases/latest/download/SupplyCall.exe`. | n |
| Runner | `windows-latest`, rodando o `build.ps1` existente (generate, vet, testes, build) | Reaproveita o script; os testes usam APIs do Windows. | n |
| Branch | Só roda a partir de `main` | Release de branch solta não teria commit estável. | n |
| Versão no menu da bandeja | Nenhuma mudança em `tray.go`/`main.go`: o menu e o `-version` leem o mesmo `main.version` que o `build.ps1` grava; o workflow confere o `-version` | Uma fonte só para a versão; conferir o `-version` prova o valor que a bandeja mostra. | y |
| Onde fica a conta da versão | Função Go pura em `internal/release`, chamada por `cmd/nextversion` | Testável com `go test`, igual ao resto do repo (AD-002). | n |

**Open questions:** none - all resolved or logged above.

---

## User Stories

### P1: Publicar uma versão com um clique ⭐ MVP

**User Story**: Como mantenedor, quero disparar a action escolhendo o bump para gerar tag, Release e `.exe` sem compilar na minha máquina.

**Why P1**: É o pedido inteiro.

**Acceptance Criteria**:

1. WHEN não existe nenhuma tag `vX.Y.Z` THEN o cálculo de versão SHALL retornar `v1.0.0` para qualquer bump.  <!-- REL-01 -->
2. WHEN a maior tag é `vA.B.C` e o bump é `patch` THEN o cálculo SHALL retornar `vA.B.(C+1)`.  <!-- REL-02 -->
3. WHEN a maior tag é `vA.B.C` e o bump é `minor` THEN o cálculo SHALL retornar `vA.(B+1).0`.  <!-- REL-03 -->
4. WHEN a maior tag é `vA.B.C` e o bump é `major` THEN o cálculo SHALL retornar `v(A+1).0.0`.  <!-- REL-04 -->
5. The cálculo SHALL escolher a maior tag por ordem numérica (`v1.10.0` > `v1.9.0`), não pela ordem de texto.  <!-- REL-05 -->
6. IF uma tag não segue `vMAJOR.MINOR.PATCH` exato THEN o cálculo SHALL ignorá-la.  <!-- REL-06 -->
7. IF o bump não é `patch`, `minor` nem `major` THEN o cálculo SHALL falhar com erro e o workflow SHALL parar antes de criar tag.  <!-- REL-07 -->
8. WHEN o workflow roda THEN ele SHALL compilar com `build.ps1 <versão sem v>`, gravando essa versão em `main.version`.  <!-- REL-08 -->
9. IF generate, vet, testes ou build falham THEN o workflow SHALL terminar sem criar tag nem Release.  <!-- REL-09 -->
10. WHEN o build passa THEN o workflow SHALL criar a tag `vX.Y.Z` no SHA do commit compilado (`github.sha`) e um Release público com esse nome.  <!-- REL-10 -->
11. The notas do Release SHALL conter o SHA completo do commit e um link para ele.  <!-- REL-11 -->
12. The Release SHALL ter `SupplyCall.exe` como asset.  <!-- REL-12 -->
13. The workflow SHALL enviar `SupplyCall.exe` também como artifact do run, com nome `SupplyCall-vX.Y.Z`.  <!-- REL-13 -->
14. IF o workflow é disparado de outra branch que não `main` THEN ele SHALL falhar antes do build.  <!-- REL-14 -->
15. WHILE um run de release está em andamento o workflow SHALL enfileirar outro disparo, sem rodar os dois ao mesmo tempo.  <!-- REL-15 -->
16. The item "Versão ..." do menu da bandeja SHALL mostrar `X.Y.Z` (a versão do Release, sem `v`), a mesma que `SupplyCall.exe -version` imprime como `Supply Call X.Y.Z`.  <!-- REL-18 -->
17. IF `SupplyCall.exe -version` não imprime `Supply Call X.Y.Z` THEN o workflow SHALL falhar sem criar tag nem Release.  <!-- REL-19 -->

**Independent Test**: `go test ./internal/release` cobre 1–7. Disparar o workflow no GitHub e ver o Release `v1.0.0` com o `.exe`, a tag no commit e o SHA nas notas.

---

### P2: Link de download no README

**User Story**: Como jogador, quero um link direto no README para baixar o `.exe` sem procurar a página de Releases.

**Why P2**: O Release já é público; o link só encurta o caminho.

**Acceptance Criteria**:

1. The README SHALL ter um link para `https://github.com/paeeglee/supply-call/releases/latest/download/SupplyCall.exe`.  <!-- REL-16 -->

**Independent Test**: Abrir o link depois do primeiro Release e o download começar.

---

## Edge Cases

- IF a tag calculada já existe no remoto THEN o workflow SHALL falhar ao criar o Release, sem sobrescrever o Release existente.  <!-- REL-17 -->
- WHEN existem tags `v1.0.0` e `v1.0.0-beta` THEN o cálculo SHALL considerar só `v1.0.0`.  <!-- coberto por REL-06 -->

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| -------------- | ----- | ----- | ------ |
| REL-01 | P1: Publicar uma versão | Execute | Implementing |
| REL-02 | P1: Publicar uma versão | Execute | Implementing |
| REL-03 | P1: Publicar uma versão | Execute | Implementing |
| REL-04 | P1: Publicar uma versão | Execute | Implementing |
| REL-05 | P1: Publicar uma versão | Execute | Implementing |
| REL-06 | P1: Publicar uma versão | Execute | Implementing |
| REL-07 | P1: Publicar uma versão | Execute | Implementing |
| REL-08 | P1: Publicar uma versão | Execute | Pending |
| REL-09 | P1: Publicar uma versão | Execute | Pending |
| REL-10 | P1: Publicar uma versão | Execute | Pending |
| REL-11 | P1: Publicar uma versão | Execute | Pending |
| REL-12 | P1: Publicar uma versão | Execute | Pending |
| REL-13 | P1: Publicar uma versão | Execute | Pending |
| REL-14 | P1: Publicar uma versão | Execute | Pending |
| REL-15 | P1: Publicar uma versão | Execute | Pending |
| REL-16 | P2: Link no README | Execute | Pending |
| REL-17 | P1: Publicar uma versão | Execute | Pending |
| REL-18 | P1: Publicar uma versão | Execute | Pending |
| REL-19 | P1: Publicar uma versão | Execute | Pending |

**Coverage:** 19 total, 19 mapped, 0 unmapped.

---

## Success Criteria

- [ ] O primeiro disparo publica `v1.0.0` com `SupplyCall.exe` baixável sem login.
- [ ] A tag `v1.0.0` aponta para o mesmo SHA citado nas notas do Release.
