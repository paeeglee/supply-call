# Overlay: tarefas

## Execution Protocol (MANDATORY -- do not skip)

Implement these tasks with the `tlc-spec-driven` skill: **activate it by name and follow its Execute flow and Critical Rules.** Do not search for skill files by filesystem path. The skill is the source of truth for the full flow (per-task cycle, sub-agent delegation, adequacy review, Verifier, discrimination sensor).

**If the skill cannot be activated, STOP and tell the user - do not proceed without it.**

---

**Design**: `.specs/features/overlay/design.md`
**Status**: Done

---

## Test Coverage Matrix

> Generated from codebase, project guidelines, and spec - confirm before Execute. Guidelines found: none (repositório vazio) - strong defaults applied, mais o pedido do usuário (testes de relógio, seleção de passo e carregador; `go test -race`).

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| ---------- | ------------------ | -------------------- | ---------------- | ----------- |
| Domínio (`build`, `clock`, `api.Tracker`, `overlay.BuildView`, `hotkeys.Parse`) | unit | Todos os ramos; 1:1 com os ACs; todos os edge cases listados | `internal/<pkg>/*_test.go` | `go test ./internal/<pkg>/` |
| HTTP (`api.Client`, `api.Run`, `sim`) | integration (`httptest`) | Caminho feliz + erro/timeout + formato | `internal/<pkg>/*_test.go` | `go test ./internal/<pkg>/` |
| Persistência (`config`, `build.Scan`) | integration (`t.TempDir`) | Leitura, escrita, primeira execução, arquivo inválido | `internal/<pkg>/*_test.go` | `go test ./internal/<pkg>/` |
| Estado compartilhado (`overlay.Shared`) | unit com `-race` | Acesso concorrente sem corrida | `internal/overlay/shared_test.go` | `go test -race ./internal/overlay/` |
| Janela, bandeja, atalhos (registro), som, `main.go` | none (smoke manual com print) | Build gate + roteiro manual no Done when | - | build gate |

## Gate Check Commands

> Generated from codebase - confirm before Execute.

| Gate Level | When to Use | Command |
| ---------- | ----------- | ------- |
| Quick | Tarefa com testes unit/integration de um pacote | `go test ./internal/<pkg>/` |
| Full | Tarefa que mexe em estado concorrente ou em vários pacotes | `go test -race ./...` |
| Build | Fim de fase e tarefas sem testes | `go vet ./... && go test -race ./... && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/overlay.exe .` |

> `-race` precisa de CGO e de um compilador C (gcc). Só os testes usam; o `.exe` continua `CGO_ENABLED=0`.

---

## Execution Plan

Fases em sequência; tarefas em ordem dentro de cada fase. Cada fase corresponde a uma etapa do "Como trabalhar"; ao fim de cada fase, paro e mostro o resultado.

### Phase 1: Núcleo (etapa 1)

```
T1 → T2 → T3 → T4 → T5 → T6 → T7
```

### Phase 2: Prova de conceito janela + bandeja (etapa 2)

```
T8 → T9 → T10
```

### Phase 3: Simulação e janela completa (etapa 3)

```
T11 → T12 → T13 → T14 → T15 → T16 → T17 → T18 → T19
```

### Phase 4: Bandeja completa, atalhos, posição (etapa 4a)

```
T20 → T21 → T22 → T23 → T24 → T25
```

### Phase 5: API real, extras e entrega (etapa 4b)

```
T26 → T27 → T28 → T29 → T30 → T31 → T32
```

---

## Task Breakdown

### T1: Esqueleto do projeto

**What**: `go.mod` com `go 1.25`, `.gitignore` (`bin/`), `main.go` com `var version = "dev"` e flag `--version`.
**Where**: `main.go`
**Depends on**: None
**Reuses**: -
**Requirement**: VER-01

**Done when**:

- [x] `go build ./...` e `CGO_ENABLED=0 go build .` passam
- [x] `go run . --version` imprime `SC2 Build Overlay dev`

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `chore: scaffold go module and entrypoint`

---

### T2: Formato de tempo m:ss

**What**: `ParseTime` (aceita `m:ss`, `mm:ss`; segundos 00–59; rejeita vazio, negativo, `1:5`, `abc`) e `FormatTime` (arredonda para cima nas contagens via helper `Countdown`).
**Where**: `internal/build/timefmt.go`
**Depends on**: T1
**Reuses**: -
**Requirement**: BUILD-05, STEP-03

**Done when**:

- [x] Tabela de casos válidos e inválidos passa
- [x] Gate: `go test ./internal/build/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(build): parse and format m:ss times`

---

### T3: Carregador de build

**What**: Tipos `Build/Step/Reminder`, `Parse`/`Load` com yaml.v3, ordenação estável por tempo, validação de raça/`vs`/passos/`every_seconds`, nome padrão = arquivo, chaves extras ignoradas, erros com arquivo e passo.
**Where**: `internal/build/build.go`
**Depends on**: T2
**Reuses**: `ParseTime` (T2)
**Requirement**: BUILD-01, BUILD-02, BUILD-03, BUILD-04, BUILD-06, BUILD-07

**Done when**:

- [x] Testes: build de exemplo com acentos (`Refinaria`, `2º Command Center`, `gás`), fora de ordem, empate estável, chaves extras, YAML quebrado, `time` inválido, raça inválida, passos vazios, `every_seconds: 0`, `name` vazio
- [x] Gate: `go test ./internal/build/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(build): load and validate build yaml files`

---

### T4: Timeline e passo atual

**What**: Agrupamento por tempo e `(*Timeline).At(t, warn)` com estados passado/agora/atual/aviso/futuro, contagem e `Done`.
**Where**: `internal/build/timeline.go`
**Depends on**: T3
**Reuses**: tipos do T3
**Requirement**: STEP-01, STEP-02, STEP-03, STEP-04, STEP-05, STEP-06

**Done when**:

- [x] Testes por AC + edge cases: t < 0, grupos a < 3 s um do outro, limite exato de 3 s, aviso em exatamente 5 s, fim da build, troca de build com t fixo
- [x] Gate: `go test ./internal/build/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(build): compute step states from game time`

---

### T5: Lembretes

**What**: `ReminderStatus(rs, t)` com ativo, contagem até o próximo disparo e `Flash` (1,5 s após o disparo).
**Where**: `internal/build/reminders.go`
**Depends on**: T4
**Reuses**: tipos do T3
**Requirement**: REM-01, REM-02, REM-03, REM-04

**Done when**:

- [x] Testes: t=0 (sem disparo), t=12 (flash), t=13,4 (flash), t=13,6 (sem flash), t=`until` exato, t>`until`, vários lembretes
- [x] Gate: `go test ./internal/build/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(build): compute periodic reminder state`

---

### T6: Relógio interpolado

**What**: `clock.Clock` com `now` injetável, ressincronização, interpolação pela taxa medida (0,25–8×, janela de 10 s, até +3 s sem leituras), pausa confirmada após 1,6 s sem mudança, anti-jitter de 2 s, reinício. (Revisado na T30 com dados reais.)
**Where**: `internal/clock/clock.go`
**Depends on**: T5
**Reuses**: -
**Requirement**: CLK-01, CLK-02, CLK-03, CLK-04, CLK-05

**Done when**:

- [x] Testes com relógio falso para cada AC, incluindo 4× (simulação), limite de +1 s sem leituras, queda de 1,5 s (mantém) e de 3 s (reinicia), acesso concorrente
- [x] Gate: `go test -race ./internal/clock/` (ou sem `-race` se não houver gcc)

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(clock): interpolate game clock between api readings`

---

### T7: Configuração

**What**: `Config`, `Dir()`, `Load`, `Save` atômico, `EnsureFirstRun` (cria config + `builds\` com `example.yml`), padrões, config malformado sem sobrescrever.
**Where**: `internal/config/config.go`
**Depends on**: T6
**Reuses**: `assets/example.yml` (criado aqui, embutido em T8)
**Requirement**: CFG-01, CFG-02, CFG-03, CFG-04

**Done when**:

- [x] Testes com `t.TempDir()`: primeira execução, ida e volta, padrões para campos ausentes, YAML quebrado não sobrescreve, sem temporário órfão após salvar
- [x] Gate: `go test ./internal/config/`

**Status**: ✅ Done
**Tests**: integration
**Gate**: quick
**Commit**: `feat(config): load, save and bootstrap config.yml`

---

### T8: Assets e ícone

**What**: `cmd/genicon` desenha o ícone e grava `assets/icon.ico` + `assets/icon.png`; `assets/assets.go` embute ícone e `example.yml`; `rsrc_windows_amd64.syso` gerado com `akavel/rsrc`.
**Where**: `assets/assets.go`
**Depends on**: None (Phase 1 concluída)
**Reuses**: `github.com/akavel/rsrc`
**Requirement**: PKG-01

**Done when**:

- [x] `go generate ./...` recria os arquivos
- [x] O `.exe` mostra o ícone no Explorer
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(assets): embed app icon and example build`

---

### T9: POC da janela

**What**: `overlay.Game` mínimo: transparente, sem borda, flutuante, 320 px, texto com acento, arrastável, encerra com `ebiten.Termination` ao receber pedido de saída.
**Where**: `internal/overlay/game.go`
**Depends on**: T8
**Reuses**: gofont, `text/v2`
**Requirement**: OVL-01, OVL-09

**Done when**:

- [x] Janela aparece por cima de outras, fundo semitransparente, "Aguardando partida…" legível
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(overlay): transparent always-on-top window poc`

---

### T10: POC da bandeja + Sair

**What**: `tray.Start` em goroutine com `LockOSThread` + `systray.Run`; ícone, tooltip com versão, "Versão x.y.z" desabilitado, "Sair" que fecha overlay e bandeja; `main.go` junta os dois.
**Where**: `internal/tray/tray.go`
**Depends on**: T9
**Reuses**: `assets` (T8), `overlay.Game` (T9)
**Requirement**: TRAY-01, TRAY-02, TRAY-08

**Done when**:

- [x] Clique no ícone abre o menu; "Sair" encerra o processo (sem sobras no Gerenciador de Tarefas)
- [x] Print da janela + menu mostrado ao usuário
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(tray): system tray with version and quit`

---

### T11: Cliente da API

**What**: `api.Client` com `Game`/`UI`, timeout 1 s (revisado na T30), decodificação tolerante.
**Where**: `internal/api/client.go`
**Depends on**: None (Phase 2 concluída)
**Reuses**: -
**Requirement**: API-01

**Done when**:

- [x] Testes `httptest`: JSON completo, campos extras, 500, JSON inválido, timeout
- [x] Gate: `go test ./internal/api/`

**Status**: ✅ Done
**Tests**: integration
**Gate**: quick
**Commit**: `feat(api): sc2 client api http client`

---

### T12: Máquina de estados da partida

**What**: `api.Tracker.Update` emitindo `Online/Offline/MatchStart/Reading/MatchEnd`, com regra de replay.
**Where**: `internal/api/tracker.go`
**Depends on**: T11
**Reuses**: tipos do T11
**Requirement**: API-02, API-03, API-04, API-05, API-06, API-07

**Done when**:

- [x] Testes por transição: offline→menu→carregando→partida, fim, queda de `displayTime` > 2 s, replay com e sem `show_replays`, falhas repetidas geram um só `Offline`
- [x] Gate: `go test ./internal/api/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(api): track match lifecycle from api snapshots`

---

### T13: Poller

**What**: `api.Run(ctx, …)` consulta a cada 500 ms, chama o tracker e entrega eventos; log só nas transições; para ao cancelar o contexto.
**Where**: `internal/api/poller.go`
**Depends on**: T12
**Reuses**: T11, T12
**Requirement**: API-01, API-05

**Done when**:

- [x] Teste `httptest` com intervalo curto: eventos na ordem, cancelamento encerra a goroutine
- [x] Gate: `go test -race ./internal/api/`

**Status**: ✅ Done
**Tests**: integration
**Gate**: full
**Commit**: `feat(api): poll game and ui endpoints`

---

### T14: Servidor de simulação

**What**: `sim.Sim` com ciclo menu→carregamento→partida→menu, velocidade, pausa, nova partida, `Handler()` no formato da API, `ListenAndServe`.
**Where**: `internal/sim/sim.go`
**Depends on**: T13
**Reuses**: tipos do `api` (T11)
**Requirement**: SIM-01, SIM-02, SIM-03, SIM-04, SIM-05

**Done when**:

- [x] Testes com relógio falso: cada fase do ciclo, 2×/4×, pausa congela, nova partida volta a 0; resposta decodifica com `api.Client`
- [x] Gate: `go test ./internal/sim/`

**Status**: ✅ Done
**Tests**: integration
**Gate**: quick
**Commit**: `feat(sim): fake sc2 client api server`

---

### T15: Estado compartilhado

**What**: `overlay.Shared` com mutex: build atual, status (aguardando/em partida/porta ocupada), visível, click-through, pedido de saída, relógio.
**Where**: `internal/overlay/shared.go`
**Depends on**: T14
**Reuses**: `clock` (T6)
**Requirement**: STEP-06, HK-01, HK-02, TRAY-08

**Done when**:

- [x] Teste concorrente (várias goroutines escrevendo, uma lendo) passa com `-race`
- [x] Gate: `go test -race ./internal/overlay/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: full
**Commit**: `feat(overlay): concurrency-safe shared state`

---

### T16: Modelo de exibição

**What**: `BuildView` puro: cabeçalho (relógio, nome, "Build concluída"), linhas com cor/risco/nota/contagem/"AGORA", piscar do aviso, linha de lembretes, altura total limitada, rolagem automática (grupo atual no primeiro terço) e manual.
**Where**: `internal/overlay/view.go`
**Depends on**: T15
**Reuses**: `build.Timeline`, `build.ReminderStatus`
**Requirement**: OVL-02, OVL-03, OVL-04, OVL-05, OVL-06, OVL-07, OVL-08, OVL-11, REM-02

**Done when**:

- [x] Testes por AC: textos exatos ("Supply Depot em 0:05", "AGORA", "Build concluída", "Aguardando partida…", "Nenhuma build selecionada"), cor por estado, piscar em 250 ms, altura limitada, rolagem mantém o atual no primeiro terço, rolagem manual reseta ao mudar o grupo
- [x] Gate: `go test ./internal/overlay/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(overlay): pure view model for build list`

---

### T17: Desenho completo

**What**: `Game.Draw/Update` desenha o `View` (painel com `opacity`, textos 14/11 px, risco, borda "modo mover"), ajusta altura da janela, roda do mouse.
**Where**: `internal/overlay/game.go`
**Depends on**: T16
**Reuses**: T9, T16
**Requirement**: OVL-01, OVL-03, OVL-04, OVL-05, OVL-06, OVL-09, OVL-11

**Done when**:

- [x] Com dados fixos, a janela bate com o mockup aprovado
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(overlay): render build list, header and reminders`

---

### T18: Ligação principal com poller e `-sim`

**What**: `main.go` sobe config, build selecionada, relógio, poller, `-sim`/`-sim-speed`/`-sim-length`, porta ocupada → status.
**Where**: `main.go`
**Depends on**: T17
**Reuses**: T6, T13, T14, T15
**Requirement**: SIM-01, SIM-06, API-03, API-04

**Done when**:

- [x] `overlay.exe -sim -sim-speed 4` roda uma partida inteira e reinicia sozinho
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat: wire poller, clock and simulator into overlay`

---

### T19: Submenu de simulação

**What**: Submenu "Simulação" (Pausar/Retomar, 1x/2x/4x, Nova partida), só com `-sim`.
**Where**: `internal/tray/tray.go`
**Depends on**: T18
**Reuses**: `sim.Sim` (T14)
**Requirement**: SIM-04, SIM-05

**Done when**:

- [x] Pausar congela o relógio do overlay; Nova partida volta a 0:00
- [x] Print ao usuário (fim da etapa 3)
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(tray): simulation controls submenu`

---

### T20: Catálogo da pasta de builds

**What**: `build.Scan(dir)` lista `.yml`/`.yaml` ordenados com nome exibido, build ou erro; log do motivo.
**Where**: `internal/build/catalog.go`
**Depends on**: None (Phase 3 concluída)
**Reuses**: `build.Load` (T3)
**Requirement**: TRAY-03, TRAY-05, TRAY-06

**Done when**:

- [x] Testes com `t.TempDir()`: pasta vazia, inexistente, mistura válido/inválido, `.yaml`, outros arquivos ignorados, nome vazio usa arquivo
- [x] Gate: `go test ./internal/build/`

**Status**: ✅ Done
**Tests**: integration
**Gate**: quick
**Commit**: `feat(build): scan builds folder`

---

### T21: Submenu de builds

**What**: Submenu "Build order" com checkbox, "(erro)" desabilitado, "Nenhuma build encontrada", troca que salva `selected_build` e aplica na hora; `Refresh` remove e recria itens.
**Where**: `internal/tray/tray.go`
**Depends on**: T20
**Reuses**: T20, `config.Save`
**Requirement**: TRAY-02, TRAY-03, TRAY-04, TRAY-05, TRAY-06, STEP-06

**Done when**:

- [x] Trocar a build no meio da simulação muda a lista mantendo o relógio
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(tray): build order submenu`

---

### T22: Escolher pasta

**What**: "Escolher pasta das builds…" com `zenity.SelectFile(Directory, Filename, Title)`; cancelado não muda nada.
**Where**: `internal/tray/tray.go`
**Depends on**: T21
**Reuses**: T21
**Requirement**: TRAY-07

**Done when**:

- [x] Escolher uma pasta atualiza o submenu e o config; cancelar não altera
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(tray): choose builds folder dialog`

---

### T23: Parser de atalhos

**What**: `hotkeys.Parse("Ctrl+Alt+H")` → modificadores + tecla.
**Where**: `internal/hotkeys/parse.go`
**Depends on**: T22
**Reuses**: `golang.design/x/hotkey` (tipos)
**Requirement**: HK-03, HK-04

**Done when**:

- [x] Testes: combinações válidas, minúsculas, espaços, F1–F24, dígitos, tecla ausente, modificador desconhecido, tecla repetida
- [x] Gate: `go test ./internal/hotkeys/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(hotkeys): parse hotkey strings`

---

### T24: Registro dos atalhos

**What**: `hotkeys.Listen` registra e chama o callback; `main.go` liga a esconder e click-through; falha → log.
**Where**: `internal/hotkeys/listen.go`
**Depends on**: T23
**Reuses**: T15, T23
**Requirement**: HK-01, HK-02, HK-04

**Done when**:

- [x] Ctrl+Alt+H esconde/mostra; Ctrl+Alt+J liga o modo mover
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(hotkeys): register global hotkeys`

---

### T25: Posição da janela

**What**: Salvar ao terminar o arraste e ao sair; ao abrir, voltar para (20, 200) se estiver fora do monitor (função pura `ClampPosition` testada).
**Where**: `internal/overlay/position.go`
**Depends on**: T24
**Reuses**: `config.Save`
**Requirement**: OVL-10

**Done when**:

- [x] Testes de `ClampPosition`; posição sobrevive a reiniciar o app
- [x] Gate: `go test ./internal/overlay/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(overlay): persist window position`

---

### T26: Log e `--version`

**What**: Log em `%APPDATA%\SC2BuildOverlay\overlay.log` (truncado > 1 MB), `recover` nas goroutines, `--version` com `AttachConsole`.
**Where**: `internal/config/log.go`
**Depends on**: None (Phase 4 concluída)
**Reuses**: `config.Dir`
**Requirement**: VER-02, API-05, TRAY-06

**Done when**:

- [x] Teste do truncamento; `overlay.exe --version` (build `windowsgui`) imprime no PowerShell
- [x] Gate: `go test ./internal/config/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat: log file and version output`

---

### T27: Sugestão por raça

**What**: `build.Suggest` + ligação no `MatchStart` respeitando escolha manual desde a última partida.
**Where**: `internal/build/suggest.go`
**Depends on**: T26
**Reuses**: T20
**Requirement**: AUTO-01, AUTO-02

**Done when**:

- [x] Testes: casa raça/`vs`, `Random`, nenhuma casa, jogador não achado, abreviações `Terr`/`Prot`/`Zerg`/`random`
- [x] Gate: `go test ./internal/build/`

**Status**: ✅ Done
**Tests**: unit
**Gate**: quick
**Commit**: `feat(build): suggest build by opponent race`

---

### T28: Recarga automática da pasta

**What**: Varredura a cada 2 s por data de modificação; mudança → `tray.Refresh` e recarga da build em uso.
**Where**: `internal/build/watch.go`
**Depends on**: T27
**Reuses**: T20
**Requirement**: RELOAD-01

**Done when**:

- [x] Teste com `t.TempDir()` e varredura manual (sem sleep): criar, alterar, apagar
- [x] Gate: `go test ./internal/build/`

**Status**: ✅ Done
**Tests**: integration
**Gate**: quick
**Commit**: `feat(build): reload builds folder on change`

---

### T29: Som de aviso

**What**: Bipe senoidal de 150 ms com `ebiten/v2/audio` ao entrar em "aviso", só com `sound: true`.
**Where**: `internal/overlay/sound.go`
**Depends on**: T28
**Reuses**: T16 (detecção de transição para aviso)
**Requirement**: SND-01

**Done when**:

- [x] Com `sound: true` toca uma vez por grupo; com `false`, silêncio
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `feat(overlay): optional warning beep`

---

### T30: Verificação com a API real

**What**: Rodar com o SC2 aberto (menu, partida vs IA, replay), gravar respostas reais em `internal/api/testdata/` e ajustar o cliente/tracker se o formato divergir.
**Where**: `internal/api/testdata/`
**Depends on**: T29
**Reuses**: T11, T12
**Requirement**: API-01, API-02, API-03, API-06

**Done when**:

- [x] Testes do tracker usando as respostas reais gravadas
- [x] Diferenças encontradas relatadas ao usuário
- [x] Gate: `go test ./internal/api/`

**Status**: ✅ Done
**Tests**: integration
**Gate**: quick
**Commit**: `test(api): real sc2 api fixtures`

---

### T31: README

**What**: README em português (compilar, rodar, `-sim`, formato da build com chaves, menu, "Tela cheia em janela", atalhos/administrador, limitações).
**Where**: `README.md`
**Depends on**: T30
**Reuses**: -
**Requirement**: PKG-02

**Done when**:

- [x] Todas as seções do PKG-02 presentes
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `docs: portuguese readme`

---

### T32: Script de release

**What**: `build.ps1` com `go generate`, `go vet`, testes e `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -X main.version=$v"`.
**Where**: `build.ps1`
**Depends on**: T31
**Reuses**: -
**Requirement**: PKG-01, PKG-03

**Done when**:

- [x] `./build.ps1 1.0.0` gera `bin/overlay.exe` com ícone, sem console, `--version` = 1.0.0
- [x] Gate: build

**Status**: ✅ Done
**Tests**: none
**Gate**: build
**Commit**: `build: release script`

---

## Fix Tasks (Verifier, 1ª rodada)

### F1: Testar a janela de 10 s da taxa (CLK-02)

**What**: Teste que alimenta 20 s a 1× e 12 s a 4× e exige interpolação a ~4× (mutante M2 sobreviveu).
**Where**: `internal/clock/clock_test.go`
**Status**: ✅ Done

### F2: Testar o piso de 0,25× da taxa (CLK-02)

**What**: Teste com taxa medida de 0,125× que exige 0,25× (mutante M3 sobreviveu).
**Where**: `internal/clock/clock_test.go`
**Status**: ✅ Done

### F3: Não sobrescrever config malformado ao sair (CFG-03)

**What**: Salvar a posição ao sair só se a janela se moveu nesta execução (`overlay.PositionChanged`, testado). A 1ª versão comparava com o config e falhava justo com config malformado (Verifier, rodada 2).
**Where**: `main.go`
**Status**: ✅ Done

## Phase Execution Map

```
Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5

Phase 1:  T1 → T2 → T3 → T4 → T5 → T6 → T7
Phase 2:  T8 → T9 → T10
Phase 3:  T11 → T12 → T13 → T14 → T15 → T16 → T17 → T18 → T19
Phase 4:  T20 → T21 → T22 → T23 → T24 → T25
Phase 5:  T26 → T27 → T28 → T29 → T30 → T31 → T32
```
