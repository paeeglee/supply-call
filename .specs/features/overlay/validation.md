# Overlay Validation

**Date**: 2026-09-26
**Spec**: `.specs/features/overlay/spec.md` (versão atual, com CLK-01..03, API-01, API-05, STEP-04 revisados após T30)
**Diff range**: rodada 3 `c39d769..7eb15d1` (rodada 2: `..4cb14f2`; rodada 1: `..5fa1c50`)
**Verifier**: independent sub-agent (author ≠ verifier), evidence-or-zero

---


## Rodada 3 (re-verificação após 7eb15d1, última iteração permitida)

**Result**: PASS ✅ - os 3 gaps da rodada 1 estão fechados; gate verde; sensor 17/17 nas camadas testáveis.

### O que mudou

| Commit | Mudança | Verificação |
| ------ | ------- | ----------- |
| 7eb15d1 | `internal/overlay/position.go:38-40` - `PositionChanged(openX, openY, x, y) bool { return x != openX \|\| y != openY }`; `main.go:190` - `if overlay.PositionChanged(x, y, wx, wy) { savePosition(...) }`, onde `x, y` é a posição em que a janela abriu (`main.go:179-182`: `DefaultX, DefaultY` sem posição salva, ou a salva após `ClampPosition`); `TestPositionChanged`; texto de F3 em tasks.md | ✅ |

### ACs reavaliados

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| CFG-03 | malformado: log, padrões na memória, arquivo não sobrescrito até mudança do usuário | Pacote config: `internal/config/config_test.go:128-132` (padrões e arquivo intacto). Decisão de saída: `internal/overlay/position_test.go:47` - `PositionChanged(DefaultX, DefaultY, DefaultX, DefaultY)` deve ser `false` (config malformado abre em (20, 200) e sai sem mover → não salva); `:50` e `:53` - movida (inclusive 1 px em y) → salva. Ligação: `main.go:190` usa a posição de abertura, não o config, então o caso `WindowPosition == nil` da rodada 2 não grava mais. Log: `main.go:110`. Checagem manual do autor/coordenador com o app real (config `sound: [quebrado`, sair pela bandeja sem mover → hash inalterado, log "usando padrões"): relatada, não reexecutada por este Verifier | ✅ PASS |
| TRAY-08 / OVL-10 (salvar ao sair) | posição salva ao fechar | `internal/overlay/position_test.go:50,53` - janela movida → `true`; salvar em si em `main.go:190-192` | ✅ PASS (decisão) / 🖐 ligação |
| CLK-02 | (sem mudança desde a rodada 2) | `internal/clock/clock_test.go:135,154` | ✅ PASS |

Demais ACs: nenhum código mudou fora de `internal/overlay/position.go`, `position_test.go` e `main.go:190`; resultados das rodadas 1 e 2 mantidos.

### Discrimination Sensor (rodada 3)

Scratch: `git worktree add --detach <scratchpad>/wt3 HEAD` (HEAD = 7eb15d1), uma mutação por vez, arquivo restaurado byte a byte, `git worktree remove --force`. `git status --porcelain` do repo real idêntico antes e depois (só os 3 arquivos `.specs` não rastreados deste Verifier). Sem `git stash`.

| Mutation | File:line | Description | Killed? |
| -------- | --------- | ----------- | ------- |
| M1 | `internal/clock/clock.go:70` | Qualquer leitura igual pausa | ✅ Killed |
| M2 | `internal/clock/clock.go:80-82` | Remove a janela de 10 s | ✅ Killed (`TestClockRateUsesOnlyLast10Seconds`) |
| M2b | `internal/clock/clock.go:28` | `rateWindow` 10 s → 30 s | ✅ Killed |
| M3 | `internal/clock/clock.go:117` | Remove o piso de 0,25× | ✅ Killed (`TestClockRateFloor`) |
| M3b | `internal/clock/clock.go:30` | `minRate` 0,25 → 0,1 | ✅ Killed |
| M4 | `internal/clock/clock.go:125` | Remove o teto de 3 s | ✅ Killed |
| M5 | `internal/build/timeline.go:70` | Janela "agora" `<` → `<=` | ✅ Killed |
| M6 | `internal/build/timeline.go:75-77` | Remove o aviso do grupo seguinte | ✅ Killed |
| M7 | `internal/api/tracker.go:141` | Tolera 3 falhas em vez de 2 | ✅ Killed |
| M8 | `internal/api/tracker.go:162` | Remove "queda > 2 s = partida nova" | ✅ Killed |
| M9 | `internal/overlay/view.go:170` | Atual no meio (`/4` → `/2`) | ✅ Killed |
| M10 | `internal/overlay/view.go:142` | Inverte a fase do piscar | ✅ Killed |
| M11 | `internal/build/reminders.go:30` | `<=` → `<` no último disparo | ✅ Killed |
| M12 | `internal/build/reminders.go:26` | Flash de 1,8 s | ✅ Killed |
| M14 (novo, CFG-03) | `internal/overlay/position.go:39` | `PositionChanged` sempre `true` (salva sempre ao sair) | ✅ Killed (`TestPositionChanged`) |
| M15 (novo) | `internal/overlay/position.go:39` | Ignora o eixo y (`x != openX`) | ✅ Killed |
| M16 (novo) | `internal/overlay/position.go:39` | Sempre `false` (nunca salva ao sair) | ✅ Killed |

A troca da chamada em `main.go:190` continua sem teste automatizado (`package main`, camada manual na matriz); a decisão em si agora é testada e discriminada.

**Sensor (rodada 3)**: 17/17 killed.

### Gate (rodada 3)

- `go vet ./...`: exit 0
- `CGO_ENABLED=1 go test -race -count=1 ./...`: 107 passed, 0 failed, 0 skipped (7 pacotes `ok`); 0 → 107 desde `c39d769` (+1 desde a rodada 2)
- `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o <scratchpad>/overlay.exe .`: exit 0

### Observação residual (não bloqueia)

- `internal/overlay/game.go:114,187`: `LastPosition()` só é preenchido no primeiro `Update`. Se o app sair antes do primeiro quadro, `(0,0)` difere da posição de abertura e seria salvo. Janela de tempo desprezível (sair pela bandeja antes do primeiro quadro); registrado só como nota.

### Requirement Traceability (verificado nesta validação)

A tabela de `spec.md` não foi editada (o Verifier só grava `validation.md` e os lessons); estes são os status verificados para o autor aplicar:

| Requirement | New Status |
| ----------- | ---------- |
| BUILD-01..07, CLK-01..05, STEP-01..06, API-02..07, OVL-02..05, OVL-07, OVL-08, REM-01..04, CFG-01..04, SIM-01..06, HK-03, PKG-02, PKG-03, AUTO-02 | ✅ Verified (teste automatizado) |
| API-01, OVL-06, OVL-09, OVL-11, TRAY-03, TRAY-05, TRAY-06, HK-01, HK-02, HK-04, AUTO-01, RELOAD-01 | ✅ Verified (lógica automatizada) + 🖐 ligação manual |
| OVL-01, OVL-10, TRAY-01, TRAY-02, TRAY-04, TRAY-07, TRAY-08, VER-01, VER-02, PKG-01, SND-01 | 🖐 Verified manual/smoke (autor + UAT com o jogo real); OVL-10/TRAY-08 com a decisão de salvar testada |

### Overall (rodada 3)

**Overall**: ✅ Ready

**Spec-anchored check**: 60 ACs com evidência automatizada batem o spec; 0 gaps; 0 spec-precision gaps; 11 ACs só manuais (permitido pela matriz)
**Sensor**: 17/17 mutations killed
**Gate**: 107 passed, 0 failed; vet e build `CGO_ENABLED=0` ok

---

## Rodada 2 (histórico: re-verificação após a80c4fc e 4cb14f2)

**Veredito (rodada 2)**: não pronto - Fix 1 e Fix 2 resolvidos; Fix 3 (CFG-03) não resolvia o caso que o AC descreve.

### O que mudou

| Commit | Mudança | Verificação |
| ------ | ------- | ----------- |
| a80c4fc | `TestClockRateUsesOnlyLast10Seconds` e `TestClockRateFloor` em `internal/clock/clock_test.go`; T6/T11 corrigidos e seção "Fix Tasks" em tasks.md | ✅ M2, M2b, M3, M3b mortos (abaixo) |
| 4cb14f2 | `main.go:190` - salva ao sair só se `st.get().WindowPosition` for `nil` ou diferente de `LastPosition()` | ❌ ver CFG-03 |

### ACs reavaliados

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| CLK-02 | taxa dos últimos 10 s; 1× com < 2 s; limitada a 0,25–8×; ≤ 3 s de interpolação | `internal/clock/clock_test.go:135` - `math.Abs(got-want) > 0.05` com `want := 68 + 4*(0.45+0.25)` (1× por 20 s, depois 4× por 12 s); `internal/clock/clock_test.go:154` - `!near(got, want)` com `want := 2 + 0.25*(0.45+1.0)` (taxa medida 0,125×); mais os asserts da rodada 1 (`:80,102,113,122,137`) | ✅ PASS |
| CFG-03 | malformado: padrões na memória, **arquivo não sobrescrito até uma mudança do usuário** | Pacote: `internal/config/config_test.go:128-132` continua ✅. Ligação: `main.go:190` - `if p := st.get().WindowPosition; p == nil \|\| p.X != wx \|\| p.Y != wy { savePosition(...) }`. Com config malformado, `EnsureFirstRun` devolve `Defaults(dir)` (`internal/config/config.go:67`), cujo `WindowPosition` é `nil` (`internal/config/config.go:50-57`). Logo `p == nil` é verdadeiro e a saída **continua gravando os padrões sobre o arquivo malformado**, exatamente o cenário do AC. A correção só evita a gravação quando o config é válido e já tem posição. | ❌ Needs Fix (Fix 3b) |

Todos os outros ACs: sem mudança de código fora de `main.go:187-192` e `clock_test.go`; resultados da rodada 1 mantidos.

### Discrimination Sensor (rodada 2)

Scratch: `git worktree add --detach <scratchpad>/wt2 HEAD` (HEAD = 4cb14f2), uma mutação por vez, arquivo restaurado byte a byte, `git worktree remove --force`. `git status --porcelain` do repo real idêntico antes e depois (só os 3 arquivos `.specs` não rastreados deste Verifier).

| Mutation | File:line | Description | Killed? |
| -------- | --------- | ----------- | ------- |
| M1 | `internal/clock/clock.go:70` | Qualquer leitura igual pausa | ✅ Killed |
| M2 | `internal/clock/clock.go:80-82` | Remove o descarte de mudanças com mais de 10 s | ✅ Killed (`TestClockRateUsesOnlyLast10Seconds`) |
| M2b (novo) | `internal/clock/clock.go:28` | `rateWindow` 10 s → 30 s | ✅ Killed (`TestClockRateUsesOnlyLast10Seconds`) |
| M3 | `internal/clock/clock.go:117` | Remove o piso de 0,25× | ✅ Killed (`TestClockRateFloor`) |
| M3b (novo) | `internal/clock/clock.go:30` | `minRate` 0,25 → 0,1 | ✅ Killed (`TestClockRateFloor`) |
| M4 | `internal/clock/clock.go:125` | Remove o teto de 3 s | ✅ Killed |
| M5 | `internal/build/timeline.go:70` | Janela "agora" `<` → `<=` | ✅ Killed |
| M6 | `internal/build/timeline.go:75-77` | Remove o aviso do grupo seguinte | ✅ Killed |
| M7 | `internal/api/tracker.go:141` | Tolera 3 falhas em vez de 2 | ✅ Killed |
| M8 | `internal/api/tracker.go:162` | Remove "queda > 2 s = partida nova" | ✅ Killed |
| M9 | `internal/overlay/view.go:170` | Atual no meio (`/4` → `/2`) | ✅ Killed |
| M10 | `internal/overlay/view.go:142` | Inverte a fase do piscar | ✅ Killed |
| M11 | `internal/build/reminders.go:30` | `<=` → `<` no último disparo | ✅ Killed |
| M12 | `internal/build/reminders.go:26` | Flash de 1,8 s | ✅ Killed |
| M13 (CFG-03) | `main.go:190` | Condição de saída → `if true` (salva sempre) | 🖐 Não testável: `package main` não tem testes (camada "none (smoke manual)" na matriz); `go test ./...` passa. Não conta no sensor. |

**Sensor (rodada 2)**: 14/14 killed nas camadas testáveis; M13 manual.

### Gate (rodada 2)

- `go vet ./...`: exit 0
- `CGO_ENABLED=1 go test -race -count=1 ./...`: 106 passed, 0 failed, 0 skipped (7 pacotes `ok`); +2 testes em relação à rodada 1
- `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o <scratchpad>/overlay.exe .`: exit 0

### Fix 3b: CFG-03 continua sobrescrevendo config malformado ao sair

- **Root cause**: `main.go:190` compara com `cfg.WindowPosition`, que é `nil` justamente quando o config está malformado (padrões em memória). `p == nil` → salva.
- **Fix task**: comparar com a posição em que a janela foi aberta (`x, y` de `main.go:179-182`, que vale `DefaultX, DefaultY` quando não há posição salva) em vez de `cfg.WindowPosition`, ou usar uma flag "config carregado com erro e sem mudança do usuário". Para tornar testável, extrair a decisão para uma função pura (ex.: `shouldSaveOnExit(start, last image.Point) bool`) com teste. Verificação manual: `config.yml` com YAML quebrado, abrir, sair pelo menu sem arrastar, arquivo byte a byte idêntico; depois arrastar e sair, arquivo regravado.
- **Priority**: Minor (perda do arquivo quebrado do usuário), mas o AC CFG-03 segue violado.
- **Iteração**: 2 de 3 do ciclo fix → re-verify.

### Overall (rodada 2)

**Overall (rodada 2)**: ❌ Not Ready - CLK-02 ✅ resolvido (14/14 mutantes mortos); CFG-03 ❌ ainda violado em `main.go:190`.

---

## Rodada 1 (histórico)

## Task Completion

| Task | Status | Notes |
| ---- | ------ | ----- |
| T1–T32 | ✅ Done | Todas marcadas ✅ em tasks.md, todos os "Done when" marcados `[x]`. |
| T6 / T11 (texto) | ⚠️ Doc drift | O "What" de T6 ainda diz "0–8×, até +1 s" e o de T11 "timeout 400 ms"; o código e o spec atual usam 0,25–8×, 3 s e 1 s. Só texto de tarefa, sem efeito no código. |

---

## Spec-Anchored Acceptance Criteria

Legenda: ✅ PASS = asserção mira o resultado exato do spec · ❌ GAP = sem evidência automatizada numa camada que exige teste · 🖐 Manual = AC que vive só em camada "none (smoke manual)" da matriz (janela, bandeja, registro de atalhos, som, `main.go`), verificado pelo autor com prints/mensagens Win32 sintetizadas e UAT com o jogo real; sem evidência automatizada, não conta como falha.

### P1: Carregar build

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| BUILD-01 | name/race/vs/steps(time s, action, note)/reminders idênticos, com acentos | `internal/build/build_test.go:21` - `s.Time != 48 \|\| s.Action != "Refinaria" \|\| s.Note != "3 SCVs no gás quando terminar"`; `:24` - `Action != "2º Command Center"`; `:31` - `r.Text != "SCV!" \|\| r.EverySeconds != 12 \|\| r.Until != 420`; `vs` em `:130` - `b.Race != race \|\| b.Vs != race` | ✅ PASS |
| BUILD-02 | ordenado por time, estável | `internal/build/build_test.go:56` - `strings.Join(got, ",") != "A,B,C,D"` | ✅ PASS |
| BUILD-03 | chaves desconhecidas ignoradas sem erro | `internal/build/build_test.go:72` - `t.Fatalf("unknown keys must be ignored: %v", err)` | ✅ PASS |
| BUILD-04 | erro com nome do arquivo e motivo | `internal/build/build_test.go:82` - `!strings.Contains(err.Error(), "quebrado.yml")` | ✅ PASS |
| BUILD-05 | erro citando passo e valor; seg 00–59 | `internal/build/build_test.go:97-98` - contém `"b.yml","passo 2","1:5"`; `internal/build/timefmt_test.go:29` - `"1:60"` rejeitado | ✅ PASS |
| BUILD-06 | erro de validação p/ race/vs inválidos, steps vazio, every_seconds ≤ 0 | `internal/build/build_test.go:114-116` - `err == nil` → falha, erro nomeia `v.yml` | ✅ PASS |
| BUILD-07 | nome vazio → arquivo sem extensão | `internal/build/build_test.go:141` - `DisplayName() != "ling_bane"`; usado no cabeçalho em `internal/overlay/view.go:85` | ✅ PASS |

### P1: Relógio

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| CLK-01 | exibe D + taxa×0,45; mantém exibido se < 2 s abaixo | `internal/clock/clock_test.go:62` - `near(got, 42.45)`; `:179` - `near(got, 15.85)` (held) | ✅ PASS |
| CLK-02 | taxa pelas mudanças dos **últimos 10 s**; 1× com < 2 s; **limitada a 0,25–8×**; interpolação ≤ 3 s | `internal/clock/clock_test.go:80,102,113,122,137` cobrem interpolação, 4×, 1× com histórico curto, teto 8×, teto de 3 s. **Janela de 10 s e piso 0,25× sem asserção**: mutantes M2 e M3 sobreviveram | ❌ GAP (parcial) |
| CLK-03 | mesmo valor > 1,6 s após a mudança → para exatamente no valor; leituras faltando não pausam | `internal/clock/clock_test.go:155` - `near(got, 14)` após +1,7 s; `:150` ainda correndo em +1,5 s; `:133` - sem leituras continua correndo | ✅ PASS |
| CLK-04 | queda > 2 s → partida nova, novo valor | `internal/clock/clock_test.go:195` - `near(got, 0.45)` | ✅ PASS |
| CLK-05 | fonte de tempo injetada | `internal/clock/clock_test.go:33` - `New(ft.Now)`; todos os testes do pacote usam `fakeTime` | ✅ PASS |

### P1: Passo atual

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| STEP-01 | mesmo time = um grupo | `internal/build/timeline_test.go:29` - `len(g.Steps) != 2 \|\| g.Steps[0].Action != "Refinaria"` | ✅ PASS |
| STEP-02 | `[time,time+3)` = agora, anteriores passados; posterior vencido tira o "agora" | `internal/build/timeline_test.go:73` (casos `:63-65`: 37,9 → Now; 38 → Passed); `:89-91` - `At(234) == [Passed Now]` | ✅ PASS |
| STEP-03 | primeiro `time > t` é o atual, contagem arredondada p/ cima | `internal/build/timeline_test.go:73` (caso `:60` - 29,2 → countdown 6); `internal/build/timefmt_test.go:47` | ✅ PASS |
| STEP-04 | atual (ou grupo após "agora") com contagem ≤ warning_seconds → aviso | `internal/build/timeline_test.go:73` (caso `:61` - exatamente 5 s → Warning); `:94-95` - `At(233) == [Now Warning]`; `:101` - warn=10 | ✅ PASS |
| STEP-05 | t ≥ último+3 → todos passados, Done | `internal/build/timeline_test.go:73` (caso `:68` - 69 → all Passed, done=true) | ✅ PASS |
| STEP-06 | troca de build recalcula sem reiniciar o relógio | `internal/build/timeline_test.go:118`; `internal/overlay/shared_test.go:83` - `approx(got, 2.7)` após `SetBuild` | ✅ PASS |

### P1: Detecção de partida

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| API-01 | /game e /ui em paralelo, 500 ms, timeout 1 s | `internal/api/client_test.go:88` - `Timeout != time.Second`. Paralelismo (`internal/api/poller.go:196-201`) e 500 ms (`main.go:32`) sem asserção automatizada | ✅ PASS (timeout) / 🖐 intervalo e paralelismo |
| API-02 | activeScreens vazio + players → em partida | `internal/api/tracker_test.go:45`; `:70` sem players → None; `internal/api/real_test.go:68` com respostas reais | ✅ PASS |
| API-03 | entrada em partida ou queda > 2 s → partida nova | `internal/api/tracker_test.go:63-64` (1,5 s = Reading; queda p/ 3 = MatchStart); `internal/overlay/shared_test.go:93` | ✅ PASS |
| API-04 | sair → "Aguardando partida…" e lista em 0:00 | `internal/api/tracker_test.go:48`; `internal/overlay/shared_test.go:88` - `Status != Waiting \|\| Time != 0`; `internal/overlay/view_test.go:62` - `Status[0] != "Aguardando partida…"` | ✅ PASS |
| API-05 | 3 falhas seguidas → aguardando; log só nas transições; 1–2 falhas ignoradas | `internal/api/tracker_test.go:51-53` (3ª falha = MatchEnd+WentOffline), `:99-104`; `internal/api/poller_test.go:82` - exatamente 1 `WentOffline` | ✅ PASS |
| API-06 | replay fora de partida por padrão | `internal/api/tracker_test.go:78-79` - kind None | ✅ PASS |
| API-07 | replay como partida com show_replays | `internal/api/tracker_test.go:89-90` - MatchStart/Reading | ✅ PASS |

### P1: Janela

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| OVL-01 | sem borda, por cima, transparente, 320 px, 14 px | `internal/overlay/view_test.go:108` - `l.Size != 14`; resto em `internal/overlay/game.go:22,93-104` | 🖐 Manual (+ tamanho do texto ✅) |
| OVL-02 | relógio m:ss + nome; "Build concluída" | `internal/overlay/view_test.go:54` - `Clock != "3:04" \|\| Title != "Terran Bio"`; `:166` - `"Build concluída"` | ✅ PASS |
| OVL-03 | passados cinza riscados; atual amarelo "Ação em m:ss"; agora laranja "AGORA"; futuros brancos com tempo | `internal/overlay/view_test.go:108` (tabela `:97-101`); `:126` - `"Supply Depot — AGORA"`, `ColorAlert` | ✅ PASS |
| OVL-04 | aviso alterna laranja/amarelo a cada 250 ms | `internal/overlay/view_test.go:140` - Alert/Current/Alert em 0/250/500 ms | ✅ PASS |
| OVL-05 | nota embaixo, 11 px | `internal/overlay/view_test.go:158` - `n.Size != 11`, `n.Y > step.Y` | ✅ PASS |
| OVL-06 | altura cabe a build, limitada ao monitor − 80 | `internal/overlay/view_test.go:222,242`; o "− 80" está em `internal/overlay/game.go:25,126` | ✅ PASS (modelo) / 🖐 −80 px |
| OVL-07 | rola p/ manter o atual no primeiro terço, sem remover passados | `internal/overlay/view_test.go:251` - `pos > visible/3` falha; `:254` - 60 linhas mantidas | ✅ PASS |
| OVL-08 | "Nenhuma build selecionada" mantendo o relógio | `internal/overlay/view_test.go:81` - `Clock != "1:05" \|\| Status[0] != "Nenhuma build selecionada"` | ✅ PASS |
| OVL-09 | arrastável + borda amarela "modo mover" com click-through off | `internal/overlay/view_test.go:213` - `MoveMode`; arraste/borda em `internal/overlay/game.go:167-183,255-260` | ✅ PASS (flag) / 🖐 arraste e borda |
| OVL-10 | posição salva no fim do arraste e ao fechar | `main.go:177,187-188` | 🖐 Manual |
| OVL-11 | roda rola manualmente até o grupo mudar | `internal/overlay/view_test.go:279-289` (override + clamp); reset por troca de grupo em `internal/overlay/game.go:128-131` | ✅ PASS (override) / 🖐 reset |

### P1: Lembretes

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| REM-01 | dispara em múltiplos positivos ≤ until; pisca 1,5 s | `internal/build/reminders_test.go:30` (casos `:13` t=0 sem flash, `:15` t=12 flash, `:16-17` 13,4/13,6); `internal/overlay/view_test.go:189` | ✅ PASS |
| REM-02 | linha mostra texto + contagem | `internal/overlay/view_test.go:185` - `Reminders != "SCV! 0:07"`; `:202` - separador `" · "` | ✅ PASS |
| REM-03 | t > until de todos → linha some | `internal/build/reminders_test.go:37`; `internal/overlay/view_test.go:193` - `Reminders != ""` falha | ✅ PASS |
| REM-04 | continuam com build concluída | `internal/build/reminders_test.go:68`; `internal/overlay/view_test.go:177` | ✅ PASS |

### P1: Bandeja

| AC | Evidência | Result |
| -- | --------- | ------ |
| TRAY-01, TRAY-02, TRAY-04, TRAY-07, TRAY-08 | `internal/tray/tray.go:81-107`, `main.go:154-196` (camada sem testes pela matriz) | 🖐 Manual |
| TRAY-03 | rótulos: `internal/build/catalog_test.go:39` - `"A_tvt.yml=TvT Bio"`, `"sem_nome.yml=sem_nome"`; checkbox em `internal/tray/tray.go:178` | ✅ PASS (rótulos) / 🖐 menu |
| TRAY-05 | pasta vazia/inexistente → 0 entradas: `internal/build/catalog_test.go:57-60`; texto "Nenhuma build encontrada" em `internal/tray/tray.go:171` | ✅ PASS (dados) / 🖐 menu |
| TRAY-06 | `internal/build/catalog_test.go:36,45` - `"quebrado.yml (erro)"`, erro nomeia o arquivo; log em `main.go:312-319` | ✅ PASS (rótulo) / 🖐 log |

### P1: Configuração

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| CFG-01 | `%APPDATA%\SC2BuildOverlay\config.yml` com todas as chaves | `internal/config/config_test.go:17-21` (Dir); `:52` - todas as 11 chaves no YAML | ✅ PASS |
| CFG-02 | padrões + pasta builds + exemplo selecionado | `internal/config/config_test.go:39` - `c != want` (5, 0.7, false, Ctrl+Alt+H/J, `exemplo.yml`); `:43` - exemplo gravado | ✅ PASS |
| CFG-03 | malformado: log, padrões em memória, não sobrescrever até mudança do usuário | `internal/config/config_test.go:128-132` - padrões e arquivo intacto após `EnsureFirstRun`. **Porém** `main.go:187-188` chama `savePosition` sempre ao sair, gravando os padrões sobre o arquivo malformado mesmo sem arraste nem troca de build | ⚠️ PASS no pacote, falha na ligação (ver Fix 3) |
| CFG-04 | gravação atômica (temp + rename) | `internal/config/config_test.go:145` - só `config.yml` resta após salvar duas vezes; `internal/config/config.go:81-94` | ✅ PASS |

### P1: Simulação

| AC | Spec-defined outcome | `file:line` + assertion | Result |
| -- | -------------------- | ----------------------- | ------ |
| SIM-01 | /game e /ui no formato real em 6119 | `internal/sim/sim_test.go:65` (decodifica com `api.Client`); `:170` servidor real; ligação em `main.go:121-127` | ✅ PASS |
| SIM-02 | menu 3 s → carregamento 3 s → partida até length → menu | `internal/sim/sim_test.go:65` (sequência `:75-87`) | ✅ PASS |
| SIM-03 | 2×/4× | `internal/sim/sim_test.go:96` - `DisplayTime != 10*speed` | ✅ PASS |
| SIM-04 | Pausar congela até Retomar | `internal/sim/sim_test.go:120,125` | ✅ PASS |
| SIM-05 | Nova partida → carregamento, 0:00 | `internal/sim/sim_test.go:136,140` | ✅ PASS |
| SIM-06 | porta ocupada → log + "Porta 6119 ocupada (jogo aberto?)" | `internal/sim/sim_test.go:153`; `internal/overlay/view_test.go:73`; `internal/overlay/shared_test.go:101` (status não é sobrescrito) | ✅ PASS |

### P2 / P3

| AC | `file:line` + assertion | Result |
| -- | ----------------------- | ------ |
| HK-01 | `internal/overlay/shared_test.go:116` - escondido ⇒ `Passthrough()`; registro em `internal/hotkeys/listen.go` | ✅ PASS (estado) / 🖐 tecla |
| HK-02 | `internal/overlay/shared_test.go:111` - `ClickThrough` alterna | ✅ PASS (estado) / 🖐 tecla |
| HK-03 | `internal/hotkeys/parse_test.go:32` (minúsculas, espaços, F1/F24, dígitos); `:51` inválidos | ✅ PASS |
| HK-04 | `internal/hotkeys/parse_test.go:51` (erro); log e segue em `main.go:149-151` | ✅ PASS (erro) / 🖐 log |
| VER-01, VER-02 | `main.go:30,94-96,285-295` | 🖐 Manual |
| PKG-01 | build gate `CGO_ENABLED=0` passou (abaixo); ícone/sem console | 🖐 Manual |
| PKG-02 | `README.md:40,59,101,137,15-25,147,158,184` - todas as seções pedidas | ✅ PASS (inspeção) |
| PKG-03 | gate `go test -race ./...` passou | ✅ PASS |
| SND-01 | `internal/overlay/sound.go:12` (150 ms), `internal/overlay/game.go:154-165` | 🖐 Manual |
| AUTO-01 | `internal/build/suggest_test.go:37` (primeira em ordem de arquivo, abreviações da API); `internal/api/players_test.go:11`; condições "sem escolha manual" e "sem salvar" em `main.go:201-226` | ✅ PASS (seleção) / 🖐 condições |
| AUTO-02 | `internal/build/suggest_test.go:32` - nenhuma casa → `"", false`; `internal/api/players_test.go:14` - jogador não achado | ✅ PASS |
| RELOAD-01 | `internal/build/watch_test.go:22,33,43` - criar/alterar/apagar; tick de 2 s + recarga em `main.go:230-254` | ✅ PASS (detecção) / 🖐 ≤ 3 s e menu |

**Status**: ❌ 1 AC com gap automatizado (CLK-02, parcial) + 1 falha de ligação (CFG-03 em `main.go`). Nenhum spec-precision gap: todos os ACs com teste têm resultado preciso no spec.

### ACs só com verificação manual/smoke (sem evidência automatizada, permitido pela matriz)

OVL-01 (janela), OVL-06 (−80 px), OVL-09 (arraste/borda), OVL-10, OVL-11 (reset), TRAY-01, TRAY-02, TRAY-04, TRAY-07, TRAY-08, partes de menu de TRAY-03/05/06, HK-01/02 (registro), HK-04 (log), VER-01, VER-02, PKG-01, SND-01, API-01 (intervalo/paralelismo), AUTO-01 (condições em `main.go`), RELOAD-01 (tick/menu). O autor verificou com prints e mensagens Win32 sintetizadas; o usuário fez UAT com o jogo real.

---

## Discrimination Sensor

Scratch: `git worktree add --detach <scratchpad>/wt HEAD`, uma mutação por vez, `go test -count=1 <pkg>`, restauração do arquivo, `git worktree remove --force`. `git status --porcelain` do repo real: vazio antes e depois (idêntico).

| Mutation | File:line | Description | Killed? |
| -------- | --------- | ----------- | ------- |
| M1 | `internal/clock/clock.go:70` | Qualquer leitura igual pausa (remove o limiar de 1,6 s) | ✅ Killed (3 testes) |
| M2 | `internal/clock/clock.go:80-82` | Remove o descarte das mudanças com mais de 10 s (taxa medida na partida toda) | ❌ Survived |
| M3 | `internal/clock/clock.go:117` | Remove o piso de 0,25× da taxa | ❌ Survived |
| M4 | `internal/clock/clock.go:125` | Remove o teto de 3 s de interpolação | ✅ Killed |
| M5 | `internal/build/timeline.go:70` | Janela "agora" `t <` → `t <=` time+3 | ✅ Killed |
| M6 | `internal/build/timeline.go:75-77` | Remove o aviso do grupo seguinte durante "agora" | ✅ Killed |
| M7 | `internal/api/tracker.go:141` | Tolera 3 falhas em vez de 2 (`<` → `<=`) | ✅ Killed |
| M8 | `internal/api/tracker.go:162` | Remove "queda > 2 s = partida nova" | ✅ Killed |
| M9 | `internal/overlay/view.go:170` | Atual no meio da área visível (`/4` → `/2`) | ✅ Killed |
| M10 | `internal/overlay/view.go:142` | Inverte a fase do piscar | ✅ Killed |
| M11 | `internal/build/reminders.go:30` | Último disparo em `until` sem contagem (`<=` → `<`) | ✅ Killed |
| M12 | `internal/build/reminders.go:26` | Flash dura 1,8 s em vez de 1,5 s | ✅ Killed |

**Sensor depth**: lightweight ampliado (12 mutações; relógio, timeline, tracker, view, lembretes)
**Sensor (rodada 1)**: 10/12 killed - 2 mutantes sobreviventes em CLK-02

---

## Interactive UAT Results

Pulado aqui: o usuário já fez UAT manual com o jogo real (etapa T30 e depois), e o autor fez smoke da janela/bandeja/atalhos com prints e mensagens Win32 sintetizadas.

---

## Code Quality

| Principle | Status |
| --------- | ------ |
| Minimum code | ✅ Pacotes pequenos, funções puras onde há lógica (`build`, `clock`, `api.Tracker`, `overlay.BuildView`). |
| Surgical changes | ✅ Repositório novo; todo arquivo do diff pertence à feature. |
| No scope creep | ✅ `-choose-folder` (processo auxiliar do seletor de pasta) e `real_match.txt` servem a ACs existentes (TRAY-07, CLK-*). |
| Matches patterns | ✅ Estilo Go idiomático, erros em português consistentes com o spec. |
| Spec-anchored outcome check (asserted values match spec) | ✅ Textos exatos ("Aguardando partida…", "Build concluída", "— AGORA", "SCV! 0:07", "(erro)"), valores exatos de relógio (42,45; 15,85; 14). |
| Per-layer Coverage Expectation met (domain 1:1 ACs; HTTP happy+error+timeout) | ❌ Domínio `clock`: CLK-02 sem asserção para janela de 10 s e piso 0,25× (M2, M3). HTTP: ✅ (feliz, 500, JSON inválido, timeout, servidor fora). |
| Every test maps to a spec requirement - no unclaimed tests | ✅ `TestOpenLog*` → T26/log (API-05, TRAY-06); `TestClampPosition*` → edge case de posição; `TestClockRealMatchRecording` → CLK-01..03. |
| Documented guidelines followed: none (tasks.md: "Guidelines found: none") - strong defaults applied | ✅ |

Observações menores (não bloqueiam):

- `main.go:187-188`: posição salva sempre ao sair, mesmo sem mudança (ver Fix 3).
- `internal/build/suggest.go:120`: `NormalizeAPIRace` aceita qualquer prefixo (`"T"` → Terran, `"P"` → Protoss). Inofensivo com a API real.
- `internal/sim/sim.go:62-68`: um único `dt` que atravessa duas fases só avança uma. Com o poller a 500 ms não aparece.
- `internal/tray/tray.go:96-98`: com `-sim`, "Simulação" entra entre "Escolher pasta…" e o separador; a ordem de TRAY-02 vale sem `-sim`.

---

## Edge Cases

- [x] Build selecionada apagada/inválida → "Nenhuma build selecionada": `internal/overlay/view_test.go:81` + `main.go:256-268` (manual) + rótulo "(erro)" em `internal/build/catalog_test.go:36`.
- [x] `builds_folder` inexistente → sem entradas: `internal/build/catalog_test.go:60`; texto do menu manual (`internal/tray/tray.go:171`).
- [x] Grupos a < 3 s: `internal/build/timeline_test.go:89-95`.
- [x] t < 0 → todos futuros, primeiro atual: `internal/build/timeline_test.go:58` (caso), assert em `:73`.
- [x] `window_position` fora dos monitores → (20, 200): `internal/overlay/position_test.go:30` (casos `:20-22`).

---

## Gate Check

- **Gate command**: `go vet ./... && CGO_ENABLED=1 go test -race -count=1 ./... && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o <scratchpad>/overlay.exe .` (gcc do WinLibs no PATH; build fora do repo porque `bin/overlay.exe` pode estar em uso)
- **Gate (rodada 1)**: vet exit 0; testes 104 passed, 0 failed, 0 skipped (7 pacotes `ok`); build exit 0 (`overlay.exe`, 26,9 MB)
- **Test count before feature**: 0 (`c39d769` só tinha specs)
- **Test count after feature**: 104
- **Delta**: +104
- **Skipped tests**: nenhum
- **Failures**: nenhuma

---

## Fix Plans

### Fix 1: Janela de 10 s da taxa não é testada (CLK-02)

- **Root cause**: nenhum teste roda tempo suficiente com mudança de velocidade para que o descarte de mudanças antigas (`internal/clock/clock.go:80-82`) importe. M2 sobreviveu.
- **Fix task**: em `internal/clock/clock_test.go`, alimentar 20 s a 1× e depois 12 s a 4× (com `feed`) e afirmar que `Now()` usa taxa ≈ 4 (ex.: `Now()` 250 ms após a última mudança ≈ `D + 4*(0.45+0.25)` ± 0,05). Sem a janela, a taxa fica perto de 2 e o teste falha.
- **Priority**: Major (AC P1 explícito, lógica de maior risco)

### Fix 2: Piso de 0,25× da taxa não é testado (CLK-02)

- **Root cause**: nenhum teste com o jogo mais lento que 0,25×. M3 sobreviveu.
- **Fix task**: em `internal/clock/clock_test.go`, alimentar leituras que avançam 1 s a cada 8 s reais por ≥ 16 s (taxa 0,125) e afirmar `Now() == D + 0.25*(el+0.45)`.
- **Priority**: Minor

### Fix 3: Sair sobrescreve config malformado (CFG-03 × TRAY-08)

- **Root cause**: `main.go:187-188` chama `savePosition` sempre ao sair. Com config malformado, `st.cfg` são os padrões, e sair grava os padrões sobre o arquivo do usuário sem que ele tenha mudado nada. Contraria "não sobrescrever o arquivo até uma mudança do usuário".
- **Fix task**: em `main.go`, só salvar ao sair se a posição mudou em relação à carregada (ou se houve arraste). Na mesma linha, evita gravar o config a cada saída sem mudança. Verificação manual: config com YAML quebrado, abrir e fechar pelo menu, arquivo idêntico.
- **Priority**: Minor (perda do arquivo quebrado do usuário, que ele provavelmente queria corrigir)

---

## Requirement Traceability Update

| Requirement | Previous Status | New Status |
| ----------- | --------------- | ---------- |
| BUILD-01..07, CLK-01, CLK-03..05, STEP-01..06, API-02..07, OVL-02..05, OVL-07, OVL-08, REM-01..04, CFG-01, CFG-02, CFG-04, SIM-01..06, HK-03, PKG-02, PKG-03, AUTO-02 | Pending | ✅ Verified |
| API-01, OVL-06, OVL-09, OVL-11, TRAY-03, TRAY-05, TRAY-06, HK-01, HK-02, HK-04, AUTO-01, RELOAD-01 | Pending | ✅ Verified (lógica) + 🖐 manual (ligação) |
| OVL-01, OVL-10, TRAY-01, TRAY-02, TRAY-04, TRAY-07, TRAY-08, VER-01, VER-02, PKG-01, SND-01 | Pending | 🖐 Manual/smoke |
| CLK-02 | Pending | ❌ Needs Fix (Fix 1, Fix 2) |
| CFG-03 | Pending | ❌ Needs Fix (Fix 3) |

---

## Summary

**Overall (rodada 1)**: ⚠️ Issues - não pronto até Fix 1–3

**Spec-anchored check**: 58 ACs com evidência automatizada batem o resultado do spec; 1 gap parcial (CLK-02); 1 falha de ligação (CFG-03); 0 spec-precision gaps; 11 ACs só manuais (permitido pela matriz)
**Sensor**: 10/12 mutations killed
**Gate**: 104 passed, 0 failed; vet e build `CGO_ENABLED=0` ok

**What works**: carregador e validação de build, timeline (agora/aviso/concluída), lembretes, tracker da API (tolerância a falhas, queda > 2 s, replays), modelo de exibição (textos, cores, piscar, rolagem), config, simulador, parser de atalhos, sugestão por raça, watcher. Relógio: pausa, teto de 3 s, retenção anti-jitter, reinício e teto 8×.

**Issues found**: (1) CLK-02 janela de 10 s sem teste; (2) CLK-02 piso 0,25× sem teste; (3) `main.go` sobrescreve config malformado ao sair.

**Next steps**: implementar Fix 1–3, rodar de novo o gate e as mutações M2/M3 (devem morrer), revalidar.
