# SC2 Build Order Overlay: especificação

## Problem Statement

Treinar build orders no StarCraft II exige saber, a cada segundo, o que construir agora e o que vem depois. Olhar uma lista em papel ou em outro monitor tira a atenção do jogo. O overlay mostra a build por cima do jogo, guiado só pelo relógio da partida vindo da SC2 Client API oficial (`localhost:6119`), sem tocar no jogo.

## Goals

- [ ] O relógio do overlay fica a no máximo 0,5 s do `displayTime` do jogo durante a partida.
- [ ] A build reinicia sozinha a cada partida nova, sem nenhuma tecla.
- [ ] Editar um `.yml` basta para criar ou mudar uma build, sem recompilar.
- [ ] Um único `.exe` (`CGO_ENABLED=0`, sem console) com fonte e ícone embutidos.

## Out of Scope

| Feature | Reason |
| ------- | ------ |
| Ler memória do jogo, injetar código, automatizar entradas | Viola os termos da Blizzard (risco de ban). Única fonte: SC2 Client API. |
| Detectar o que o jogador construiu | A API não informa. O overlay é guiado só pelo tempo. |
| Marcar passos manualmente durante a partida | Pedido explícito: o overlay só mostra. |
| Suporte a tela cheia exclusiva | O Windows não compõe janelas por cima dela. Só documentar no README. |
| Editor visual de builds | A build é editada à mão no `.yml`. |
| Analisador de replay | Projeto separado. Aqui só se aceita chave extra no `.yml`. |
| macOS / Linux | Alvo único: Windows amd64. |
| Esconder a janela da barra de tarefas / Alt+Tab | Não foi pedido. Fica em Deferred Ideas. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --------------------- | -------------- | --------- | ---------- |
| Formato da API | `/game` = `{isReplay, displayTime, players:[{id,name,type,race,result}]}`, raça `Terr`/`Prot`/`Zerg`/`random`; `/ui` = `{activeScreens:[...]}`, lista vazia = em partida | Documentação pública da Blizzard; o jogo não estava aberto para conferir. Conferir na etapa 4 e ajustar. | n |
| Passo atual | O próximo grupo a vencer, com contagem regressiva. No tempo, vira "AGORA" por 3 s e depois fica riscado. | Escolha do usuário. | y |
| Atalhos padrão | Ctrl+Alt+H esconde/mostra; Ctrl+Alt+J liga/desliga click-through | Escolha do usuário. Ctrl+Alt+M, a primeira escolha, já estava registrado por outro programa no PC do usuário. F12 é reservado pelo Windows (doc do `RegisterHotKey`). | y |
| Visual | Fundo preto com opacidade `opacity` (padrão 0,7); texto branco; atual amarelo; aviso piscando laranja/amarelo a 2 Hz; passados cinza riscados | Escolha do usuário. | y |
| Identificar o jogador na API | Campo `player_name` no config; adversário = outro jogador | Escolha do usuário. | y |
| Chave da build para matchup | Chave opcional `vs` (`Terran`/`Protoss`/`Zerg`/`Random`) no `.yml` | A sugestão por raça precisa saber para qual adversário a build serve; `race` já é a raça do jogador. | n |
| Estado inicial do click-through | Ligado ao abrir (cliques passam para o jogo) | Seguro para o jogo; move-se com Ctrl+Alt+J. | n |
| Esconder | Desenhar nada + click-through forçado | O Ebitengine não tem "ocultar janela"; minimizar uma janela flutuante sem borda é instável. | n |
| Fonte | Go Regular/Go Bold (`golang.org/x/image/font/gofont`), compiladas dentro do exe | Já é dependência do Ebitengine, cobre Latin-1 (acentos, º). Os bytes TTF vêm do pacote Go, não de `//go:embed`. | n |
| Lembretes | Disparam em `every_seconds`, 2×, 3×… enquanto `t <= until`. A linha mostra "SCV! 0:07" (contagem até o próximo) e pisca 1,5 s no disparo. Vários lembretes na mesma linha, separados por " · ". Após `until` somem; sem lembretes ativos, a linha some. | "Linha fixa" única; contagem ajuda a antecipar. | n |
| Fim de partida | Ao sair da tela de jogo, volta a "Aguardando partida…" com a lista em 0:00 | A próxima partida começa pronta. | n |
| Rolagem manual | Roda do mouse rola a lista quando o click-through está desligado. A rolagem automática volta quando o passo atual muda. | Permite ver passos antigos sem brigar com a rolagem automática. | n |
| Controle da simulação | `-sim` sobe o servidor falso e o overlay no mesmo processo; submenu "Simulação" na bandeja (pausar, 1x/2x/4x, nova partida) e flag `-sim-speed` | Testar sem jogo, sem outra janela. | n |
| Tolerância a jitter | O relógio exibido nunca volta menos de 2 s; queda maior que 2 s = partida nova | Evita "tremer" na ressincronização. | n |
| Granularidade real da API | `displayTime` vem em segundos inteiros; cada consulta leva ~200–300 ms e algumas falham. Timeout 1 s; `/ui` e `/game` em paralelo; pausa detectada por tempo sem mudança (CLK-03) | Verificado com o jogo em 2026-09-26 (T30); a regra antiga (duas leituras iguais = pausa) congelaria o relógio a cada segundo. | y |

**Open questions:** none. Todas foram resolvidas ou registradas acima.

---

## User Stories

### P1: Carregar build do arquivo ⭐ MVP

**User Story**: Como jogador, quero descrever a build num `.yml` para mudar o treino sem mexer no código.

**Why P1**: Sem build não há o que mostrar.

**Acceptance Criteria**:

1. WHEN um `.yml` válido é carregado THEN o carregador SHALL retornar `name`, `race`, `vs`, `steps` (tempo em segundos, `action`, `note`) e `reminders` com os textos idênticos ao arquivo, incluindo acentos. (BUILD-01)
2. WHEN os passos estão fora de ordem no arquivo THEN o carregador SHALL retorná-los ordenados por `time`, mantendo a ordem do arquivo entre passos de mesmo `time`. (BUILD-02)
3. WHEN o arquivo tem chaves desconhecidas THEN o carregador SHALL ignorá-las sem erro. (BUILD-03)
4. IF o YAML está malformado THEN o carregador SHALL retornar um erro com o nome do arquivo e o motivo. (BUILD-04)
5. IF um `time` ou `until` não segue `m:ss` (segundos 00–59) THEN o carregador SHALL retornar um erro citando o passo e o valor. (BUILD-05)
6. IF `race` ou `vs` não é `Terran`, `Protoss`, `Zerg` ou `Random`, ou `steps` está vazio, ou `every_seconds` ≤ 0 THEN o carregador SHALL retornar um erro de validação. (BUILD-06)
7. WHEN `name` está vazio THEN o sistema SHALL exibir o nome do arquivo sem extensão como nome da build. (BUILD-07)

**Independent Test**: `go test ./internal/build` com arquivos de exemplo.

---

### P1: Relógio da partida ⭐ MVP

**User Story**: Como jogador, quero um relógio suave e fiel ao do jogo.

**Why P1**: Todo o overlay depende do tempo.

**Acceptance Criteria**:

1. WHEN chega uma leitura `displayTime` D diferente da anterior THEN o relógio SHALL passar a exibir D + taxa × 0,45 s (compensa o atraso médio da consulta), exceto se isso ficar até 2 s abaixo do exibido, caso em que SHALL manter o valor exibido até alcançá-lo. (CLK-01)
2. WHILE o jogo não está pausado o relógio SHALL interpolar com a taxa medida pelas mudanças de `displayTime` nos últimos 10 s (1× com menos de 2 s de histórico; limitada a 0,25–8×), interpolando no máximo 3 s de tempo real após a última mudança (cobre consultas que falham). (CLK-02)
3. WHEN uma leitura traz o mesmo `displayTime` mais de 1,6 s depois da última mudança THEN o relógio SHALL parar exatamente nesse valor (pausa) até o valor mudar; leituras que faltam (API lenta ou com erro) não SHALL contar como pausa. (CLK-03)
4. WHEN o `displayTime` cai mais de 2 s THEN o relógio SHALL tratar como partida nova e exibir o novo valor. (CLK-04)
5. The relógio SHALL receber a fonte de tempo real por injeção, para os testes não dependerem de tempo real. (CLK-05)

**Independent Test**: `go test ./internal/clock` com relógio falso.

---

### P1: Seleção do passo atual ⭐ MVP

**User Story**: Como jogador, quero ver o que fazer agora, o que vem e o que passou.

**Why P1**: É a função principal do overlay.

**Acceptance Criteria**:

1. The timeline SHALL agrupar passos de mesmo `time` num único grupo. (STEP-01)
2. WHEN o tempo t está em `[time, time+3)` de um grupo e nenhum grupo posterior já venceu THEN esse grupo SHALL ficar no estado "agora" e os anteriores no estado "passado". (STEP-02)
3. WHEN nenhum grupo está em "agora" THEN o primeiro grupo com `time > t` SHALL ser o atual, com contagem regressiva `time − t` arredondada para cima em segundos e formatada `m:ss`. (STEP-03)
4. WHEN o grupo atual, ou o grupo logo após um grupo em "agora", tem contagem ≤ `warning_seconds` (padrão 5) THEN ele SHALL entrar no estado "aviso". (STEP-04)
5. WHEN t ≥ `time + 3` do último grupo THEN todos os grupos SHALL estar em "passado" e a build SHALL ser marcada como concluída. (STEP-05)
6. WHEN a build é trocada no meio da partida THEN os estados SHALL ser recalculados com o tempo atual, sem reiniciar o relógio. (STEP-06)

**Independent Test**: `go test ./internal/build` com tabela de tempos.

---

### P1: Detecção de partida pela API ⭐ MVP

**User Story**: Como jogador, quero que o overlay reinicie sozinho a cada partida.

**Why P1**: Pedido explícito: nenhuma tecla durante a partida.

**Acceptance Criteria**:

1. The poller SHALL consultar `/game` e `/ui` em paralelo a cada 500 ms, com timeout de 1 s por requisição. (API-01)
2. WHEN `/ui.activeScreens` está vazio e `/game.players` não está vazio THEN o sistema SHALL considerar "em partida". (API-02)
3. WHEN o estado passa de "fora de partida" para "em partida", ou o `displayTime` cai mais de 2 s, THEN o sistema SHALL iniciar uma partida nova (relógio e build do zero). (API-03)
4. WHEN o estado sai de "em partida" THEN o overlay SHALL mostrar "Aguardando partida…" e a lista em 0:00. (API-04)
5. IF a API não responde em 3 consultas seguidas THEN o overlay SHALL mostrar "Aguardando partida…" e SHALL registrar no log apenas a transição online→offline e offline→online, não cada falha; 1 ou 2 falhas seguidas numa partida SHALL ser ignoradas (a API real às vezes passa de 1 s). (API-05)
6. WHERE `show_replays` é `false` (padrão) the sistema SHALL tratar partidas com `isReplay: true` como "fora de partida". (API-06)
7. WHERE `show_replays` é `true` the sistema SHALL tratar replays como partidas normais. (API-07)

**Independent Test**: `go test ./internal/api` com `httptest`.

---

### P1: Janela do overlay ⭐ MVP

**User Story**: Como jogador, quero ver a build por cima do jogo em "Tela cheia em janela".

**Why P1**: É a entrega visível.

**Acceptance Criteria**:

1. The janela SHALL ser sem borda, sempre por cima, com fundo transparente e painel preto com alfa `opacity`, largura 320 px e texto de 14 px. (OVL-01)
2. The topo SHALL mostrar o relógio `m:ss` e o nome da build; WHEN a build está concluída THEN SHALL mostrar "Build concluída". (OVL-02)
3. The lista SHALL mostrar todos os passos: passados em cinza e riscados; atual em amarelo com "Ação em m:ss"; "agora" em laranja com "AGORA"; futuros em branco com o tempo planejado. (OVL-03)
4. WHILE o grupo atual está em "aviso" a linha SHALL alternar entre laranja e amarelo a cada 250 ms. (OVL-04)
5. WHEN um passo tem `note` THEN a nota SHALL aparecer embaixo, em fonte de 11 px. (OVL-05)
6. The altura da janela SHALL caber a build inteira, limitada à altura do monitor menos 80 px. (OVL-06)
7. WHILE a lista é maior que a janela o overlay SHALL rolar para manter o grupo atual no primeiro terço visível, sem remover os passos passados da lista. (OVL-07)
8. WHILE não há build válida selecionada o overlay SHALL mostrar "Nenhuma build selecionada" e continuar mostrando o relógio. (OVL-08)
9. WHILE o click-through está desligado o overlay SHALL ser arrastável com o botão esquerdo e SHALL mostrar uma borda amarela com o texto "modo mover". (OVL-09)
10. WHEN o arraste termina, ou o app é fechado, THEN a posição SHALL ser salva em `window_position`. (OVL-10)
11. WHILE o click-through está desligado a roda do mouse SHALL rolar a lista manualmente até o grupo atual mudar. (OVL-11)

**Independent Test**: `go test ./internal/overlay` (modelo de exibição) + print da janela com `-sim`.

---

### P1: Lembretes periódicos ⭐ MVP

**User Story**: Como jogador, quero lembretes como "SCV!" a cada 12 s.

**Why P1**: Está no formato de build pedido.

**Acceptance Criteria**:

1. WHEN t atinge um múltiplo positivo de `every_seconds` e t ≤ `until` THEN o lembrete SHALL disparar e a linha SHALL piscar por 1,5 s. (REM-01)
2. WHILE um lembrete está ativo (t ≤ `until`) a linha fixa SHALL mostrar o texto e a contagem até o próximo disparo. (REM-02)
3. WHEN t > `until` de todos os lembretes THEN a linha de lembretes SHALL sumir. (REM-03)
4. WHILE a build está concluída os lembretes SHALL continuar até o respectivo `until`. (REM-04)

**Independent Test**: `go test ./internal/build`.

---

### P1: Bandeja e ciclo de vida ⭐ MVP

**User Story**: Como jogador, quero controlar o app pelo ícone ao lado do relógio do Windows.

**Why P1**: É a única interface de controle.

**Acceptance Criteria**:

1. WHEN o app inicia THEN o ícone SHALL aparecer na bandeja com tooltip "SC2 Build Overlay x.y.z". (TRAY-01)
2. WHEN o menu é aberto THEN ele SHALL conter, nesta ordem: "Build order" (submenu), "Escolher pasta das builds…", separador, "Versão x.y.z" desabilitado, "Sair". (TRAY-02)
3. The submenu "Build order" SHALL listar cada `.yml`/`.yaml` da pasta pelo `name` (ou nome do arquivo), com checkbox na build em uso. (TRAY-03)
4. WHEN uma build é escolhida THEN a anterior SHALL ser desmarcada, a nova carregada na hora e `selected_build` salvo no config. (TRAY-04)
5. IF a pasta não tem builds THEN o submenu SHALL mostrar um item desabilitado "Nenhuma build encontrada". (TRAY-05)
6. IF um arquivo é inválido THEN ele SHALL aparecer desabilitado como "<nome do arquivo> (erro)" e o motivo SHALL ir para o log. (TRAY-06)
7. WHEN "Escolher pasta das builds…" é confirmado THEN `builds_folder` SHALL ser salvo e o submenu recarregado; IF o diálogo é cancelado THEN nada SHALL mudar. (TRAY-07)
8. WHEN "Sair" é clicado THEN o app SHALL salvar a posição da janela, parar o poller, remover o ícone e encerrar o processo com código 0. (TRAY-08)

**Independent Test**: Rodar `-sim`, abrir o menu, trocar build, sair.

---

### P1: Configuração ⭐ MVP

**User Story**: Como jogador, quero que minhas escolhas sobrevivam entre execuções.

**Why P1**: Posição, pasta e build precisam persistir.

**Acceptance Criteria**:

1. The config SHALL ficar em `%APPDATA%\SC2BuildOverlay\config.yml` com as chaves `builds_folder`, `selected_build`, `hotkeys` (`toggle_visible`, `toggle_click_through`), `opacity`, `window_position` (`x`, `y`), `warning_seconds`, `sound`, `show_replays`, `player_name`. (CFG-01)
2. WHEN o config não existe THEN o sistema SHALL criá-lo com os padrões (`warning_seconds: 5`, `opacity: 0.7`, `sound: false`, atalhos Ctrl+Alt+H/Ctrl+Alt+J) e criar `%APPDATA%\SC2BuildOverlay\builds\` com a build de exemplo selecionada. (CFG-02)
3. IF o config está malformado THEN o sistema SHALL registrar o erro no log, usar os padrões na memória e não sobrescrever o arquivo até uma mudança do usuário. (CFG-03)
4. The config SHALL ser gravado de forma atômica (arquivo temporário + rename). (CFG-04)

**Independent Test**: `go test ./internal/config` com diretório temporário.

---

### P1: Modo simulação ⭐ MVP

**User Story**: Como desenvolvedor/jogador, quero testar sem o jogo aberto.

**Why P1**: Pedido na etapa 3; é como a janela é validada.

**Acceptance Criteria**:

1. WHEN o app roda com `-sim` THEN ele SHALL servir `/game` e `/ui` em `localhost:6119` no mesmo formato da API real e rodar o overlay contra ele. (SIM-01)
2. The simulação SHALL seguir o ciclo menu (3 s) → carregamento (3 s) → partida (até `-sim-length`, padrão 10:00) → menu, repetindo. (SIM-02)
3. WHERE `-sim-speed` é 2 ou 4 the `displayTime` SHALL avançar 2× ou 4× o tempo real. (SIM-03)
4. WHEN "Pausar" é escolhido no submenu "Simulação" THEN o `displayTime` SHALL parar até "Retomar". (SIM-04)
5. WHEN "Nova partida" é escolhido THEN a simulação SHALL ir para carregamento e começar em 0:00. (SIM-05)
6. IF a porta 6119 está ocupada THEN o app SHALL registrar no log e mostrar "Porta 6119 ocupada (jogo aberto?)" no overlay. (SIM-06)

**Independent Test**: `go test ./internal/sim` com relógio falso + rodar `overlay.exe -sim -sim-speed 4`.

---

### P2: Atalhos globais

**User Story**: Como jogador, quero esconder o overlay e ligar o modo mover sem abrir menus.

**Why P2**: O app funciona sem eles, mas mover a janela depende do click-through.

**Acceptance Criteria**:

1. WHEN `toggle_visible` (padrão Ctrl+Alt+H) é pressionado THEN o overlay SHALL alternar entre visível e escondido (sem desenho e com click-through). (HK-01)
2. WHEN `toggle_click_through` (padrão Ctrl+Alt+J) é pressionado THEN o click-through SHALL alternar. (HK-02)
3. The atalhos SHALL aceitar o formato `Mod+Mod+Tecla` com `Ctrl`, `Alt`, `Shift`, `Win` e teclas `A`–`Z`, `0`–`9`, `F1`–`F24`, sem diferenciar maiúsculas. (HK-03)
4. IF um atalho é inválido ou o registro falha (já usado por outro programa) THEN o sistema SHALL registrar no log e seguir sem ele. (HK-04)

**Independent Test**: `go test ./internal/hotkeys` (parser) + apertar as teclas com `-sim`.

---

### P2: Versão

**User Story**: Como jogador, quero saber qual versão estou rodando.

**Acceptance Criteria**:

1. The versão SHALL vir de `-ldflags "-X main.version=..."`, com padrão `dev`. (VER-01)
2. WHEN o app roda com `--version` THEN ele SHALL imprimir `SC2 Build Overlay <versão>` no console de origem e sair com código 0. (VER-02)

---

### P2: Entrega

**User Story**: Como jogador, quero um único executável e um README em português.

**Acceptance Criteria**:

1. The build `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -X main.version=1.0.0"` SHALL gerar um único `.exe` com ícone próprio, sem abrir console. (PKG-01)
2. The `README.md` SHALL cobrir, em português: compilar, rodar, editar build (lista de chaves), menu da bandeja, "Tela cheia em janela" (e o aviso sobre tela cheia exclusiva), atalhos (e se precisa de administrador) e limitações. (PKG-02)
3. The `go test -race ./...` SHALL passar. (PKG-03)

---

### P3: Som de aviso

**Acceptance Criteria**:

1. WHERE `sound` é `true` the overlay SHALL tocar um bipe curto (≤ 200 ms) quando um grupo entra em "aviso". (SND-01)

---

### P3: Sugestão por raça do adversário

**Acceptance Criteria**:

1. WHEN uma partida começa, `player_name` está definido e o usuário não escolheu build no menu desde a última partida THEN o sistema SHALL selecionar a primeira build (ordem alfabética de arquivo) com `vs` igual à raça do adversário e `race` igual à raça do jogador, sem salvar no config. (AUTO-01)
2. IF nenhuma build casa ou o jogador não é encontrado pelo nome THEN a build atual SHALL ser mantida. (AUTO-02)

---

### P3: Recarga automática da pasta

**Acceptance Criteria**:

1. WHEN um `.yml`/`.yaml` da pasta é criado, alterado ou apagado THEN em até 3 s o submenu SHALL ser atualizado e a build em uso SHALL ser recarregada, se for ela. (RELOAD-01)

---

## Edge Cases

- IF a build selecionada foi apagada ou ficou inválida THEN o overlay SHALL mostrar "Nenhuma build selecionada" e o submenu SHALL mostrar a build como "(erro)" ou sem ela.
- IF `builds_folder` não existe THEN o submenu SHALL mostrar "Nenhuma build encontrada".
- WHEN dois grupos estão a menos de 3 s um do outro THEN o anterior SHALL virar "passado" quando o posterior atinge o seu tempo (STEP-02).
- WHEN t < 0 ou a partida ainda não começou THEN todos os grupos SHALL ser futuros e o primeiro SHALL ser o atual.
- IF `window_position` fica fora de todos os monitores THEN a janela SHALL voltar para (20, 200).

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| -------------- | ----- | ----- | ------ |
| BUILD-01 | P1: Carregar build | Done | ✅ Verified |
| BUILD-02 | P1: Carregar build | Done | ✅ Verified |
| BUILD-03 | P1: Carregar build | Done | ✅ Verified |
| BUILD-04 | P1: Carregar build | Done | ✅ Verified |
| BUILD-05 | P1: Carregar build | Done | ✅ Verified |
| BUILD-06 | P1: Carregar build | Done | ✅ Verified |
| BUILD-07 | P1: Carregar build | Done | ✅ Verified |
| CLK-01 | P1: Relógio | Done | ✅ Verified |
| CLK-02 | P1: Relógio | Done | ✅ Verified |
| CLK-03 | P1: Relógio | Done | ✅ Verified |
| CLK-04 | P1: Relógio | Done | ✅ Verified |
| CLK-05 | P1: Relógio | Done | ✅ Verified |
| STEP-01 | P1: Passo atual | Done | ✅ Verified |
| STEP-02 | P1: Passo atual | Done | ✅ Verified |
| STEP-03 | P1: Passo atual | Done | ✅ Verified |
| STEP-04 | P1: Passo atual | Done | ✅ Verified |
| STEP-05 | P1: Passo atual | Done | ✅ Verified |
| STEP-06 | P1: Passo atual | Done | ✅ Verified |
| API-01 | P1: Detecção de partida | Done | ✅ Verified |
| API-02 | P1: Detecção de partida | Done | ✅ Verified |
| API-03 | P1: Detecção de partida | Done | ✅ Verified |
| API-04 | P1: Detecção de partida | Done | ✅ Verified |
| API-05 | P1: Detecção de partida | Done | ✅ Verified |
| API-06 | P1: Detecção de partida | Done | ✅ Verified |
| API-07 | P1: Detecção de partida | Done | ✅ Verified |
| OVL-01 | P1: Janela | Done | ✅ Verified (manual/smoke) |
| OVL-02 | P1: Janela | Done | ✅ Verified |
| OVL-03 | P1: Janela | Done | ✅ Verified |
| OVL-04 | P1: Janela | Done | ✅ Verified |
| OVL-05 | P1: Janela | Done | ✅ Verified |
| OVL-06 | P1: Janela | Done | ✅ Verified |
| OVL-07 | P1: Janela | Done | ✅ Verified |
| OVL-08 | P1: Janela | Done | ✅ Verified |
| OVL-09 | P1: Janela | Done | ✅ Verified |
| OVL-10 | P1: Janela | Done | ✅ Verified (manual/smoke) |
| OVL-11 | P1: Janela | Done | ✅ Verified |
| REM-01 | P1: Lembretes | Done | ✅ Verified |
| REM-02 | P1: Lembretes | Done | ✅ Verified |
| REM-03 | P1: Lembretes | Done | ✅ Verified |
| REM-04 | P1: Lembretes | Done | ✅ Verified |
| TRAY-01 | P1: Bandeja | Done | ✅ Verified (manual/smoke) |
| TRAY-02 | P1: Bandeja | Done | ✅ Verified (manual/smoke) |
| TRAY-03 | P1: Bandeja | Done | ✅ Verified |
| TRAY-04 | P1: Bandeja | Done | ✅ Verified (manual/smoke) |
| TRAY-05 | P1: Bandeja | Done | ✅ Verified |
| TRAY-06 | P1: Bandeja | Done | ✅ Verified |
| TRAY-07 | P1: Bandeja | Done | ✅ Verified (manual/smoke) |
| TRAY-08 | P1: Bandeja | Done | ✅ Verified (manual/smoke) |
| CFG-01 | P1: Configuração | Done | ✅ Verified |
| CFG-02 | P1: Configuração | Done | ✅ Verified |
| CFG-03 | P1: Configuração | Done | ✅ Verified |
| CFG-04 | P1: Configuração | Done | ✅ Verified |
| SIM-01 | P1: Simulação | Done | ✅ Verified |
| SIM-02 | P1: Simulação | Done | ✅ Verified |
| SIM-03 | P1: Simulação | Done | ✅ Verified |
| SIM-04 | P1: Simulação | Done | ✅ Verified |
| SIM-05 | P1: Simulação | Done | ✅ Verified |
| SIM-06 | P1: Simulação | Done | ✅ Verified |
| HK-01 | P2: Atalhos | Done | ✅ Verified |
| HK-02 | P2: Atalhos | Done | ✅ Verified |
| HK-03 | P2: Atalhos | Done | ✅ Verified |
| HK-04 | P2: Atalhos | Done | ✅ Verified |
| VER-01 | P2: Versão | Done | ✅ Verified (manual/smoke) |
| VER-02 | P2: Versão | Done | ✅ Verified (manual/smoke) |
| PKG-01 | P2: Entrega | Done | ✅ Verified (manual/smoke) |
| PKG-02 | P2: Entrega | Done | ✅ Verified |
| PKG-03 | P2: Entrega | Done | ✅ Verified |
| SND-01 | P3: Som | Done | ✅ Verified (manual/smoke) |
| AUTO-01 | P3: Sugestão por raça | Done | ✅ Verified |
| AUTO-02 | P3: Sugestão por raça | Done | ✅ Verified |
| RELOAD-01 | P3: Recarga automática | Done | ✅ Verified |

**Coverage:** 71 total, 71 mapped to tasks, 0 unmapped. 60 verified with automated tests, 11 manual/smoke (`validation.md`, rodada 3).

---

## Success Criteria

- [ ] Com `-sim -sim-speed 4`, uma partida de 10:00 roda do início ao fim com passos riscando no tempo certo e reinicia sozinha.
- [ ] Com o SC2 em "Tela cheia em janela", o relógio do overlay bate com o do jogo (±0,5 s) numa partida real.
- [ ] `go test -race ./...` passa; `CGO_ENABLED=0` compila.
