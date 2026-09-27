# SC2 Build Overlay

Overlay para treinar **build orders no StarCraft II**. É uma janela pequena e semitransparente, sempre por cima do jogo. Ela acompanha o relógio da partida e mostra, a cada momento, o que você precisa fazer agora, o que vem a seguir e o que já passou.

> **Plataforma:** o programa foi criado e testado **somente no Windows** (Windows 10 Pro 64 bits, com StarCraft II e monitor ultrawide 2560×1080). **Não foi testado no macOS** nem no Linux. O código usa APIs do Windows (bandeja do sistema, atalhos globais, estilo da janela), e o build só é suportado para Windows.

---

## O que é

- É um **overlay de treino**: você escolhe uma build order (um arquivo `.yml` que você mesmo edita) e o overlay vai mostrando os passos no tempo certo durante a partida.
- É **guiado só pelo tempo**. Durante a partida você não aperta nada e não marca passos: ele só mostra.
- A **única fonte de dados** é a **SC2 Client API oficial** (`http://localhost:6119`), um servidor HTTP local que o próprio jogo abre enquanto está rodando.

### O que ele **não** faz (de propósito)

- **Não** lê a memória do jogo.
- **Não** injeta nada no processo do jogo.
- **Não** envia teclas ou cliques para o jogo.
- **Não** sabe o que você construiu, porque a API não informa isso.

A janela é só uma janela comum por cima do jogo, e o objetivo é não violar os termos de uso da Blizzard.

## O que ele faz

- **Relógio da partida**, interpolado entre as consultas à API para a contagem ficar suave. Ele para quando o jogo é pausado.
- **Início e fim de partida automáticos:** a build reinicia sozinha a cada partida nova. Fora de partida, mostra "Aguardando partida…".
- **Lista com todos os passos da build:**
  - passos que já passaram ficam em cinza e riscados;
  - o passo atual fica em amarelo, com contagem ("Supply Depot em 0:05");
  - a linha pisca em laranja nos últimos 5 s;
  - na hora do passo, ele mostra "— AGORA" por 3 s;
  - os passos futuros ficam em branco, com o tempo planejado.
- **Lembretes periódicos** configuráveis. Exemplo: "TRABALHADOR" a cada 12 s até 7:00.
- **"Build concluída"** no topo quando a build acaba.
- **Ícone na bandeja do sistema:**
  - escolher a build;
  - escolher a pasta das builds;
  - ver a versão;
  - sair.
- **Atalhos globais:**
  - `Ctrl+Alt+H` esconde e mostra o overlay;
  - `Ctrl+Alt+J` liga e desliga o modo mover (click-through).
- **Janela fora da barra de tarefas e do Alt+Tab:** o controle é só pela bandeja.
- **Recarga automática:** se você editar uma build com o overlay aberto, ela é atualizada em até ~2 s.
- **Opcionais:**
  - bipe de aviso;
  - sugestão automática de build pela raça do adversário;
  - acompanhar replays.
- **Modo simulação (`-sim`):** testa sem o jogo aberto.

---

## Como usar

1. Coloque o StarCraft II em **Opções → Gráficos → Modo de exibição → "Tela cheia (em janela)"**. Em inglês é *Windowed (Fullscreen)*.
2. Abra o `overlay.exe`. Na primeira execução ele cria:
   - `%APPDATA%\SC2BuildOverlay\config.yml`: a configuração;
   - `%APPDATA%\SC2BuildOverlay\builds\`: a pasta de builds, com uma build de exemplo já selecionada;
   - `%APPDATA%\SC2BuildOverlay\overlay.log`: o log.
3. Entre numa partida. O relógio começa sozinho.
4. Para mover a janela, aperte `Ctrl+Alt+J`: aparece uma borda amarela escrita "modo mover". Arraste com o mouse e aperte `Ctrl+Alt+J` de novo. A posição fica salva.

### Modos de tela do SC2

| Modo do SC2 | Overlay |
| --- | --- |
| Tela cheia (em janela) | ✅ Aparece. É o modo recomendado. |
| Janela | ✅ Aparece, mas o jogo fica numa janela menor que a tela. |
| Tela cheia (exclusiva) | ❌ **Não aparece** (testado). Desenhar sobre a tela cheia exclusiva exigiria injetar código no jogo, e este programa não faz isso. |

### Monitor ultrawide (21:9)

O SC2 não suporta ultrawide: em "Tela cheia (em janela)" ele desenha em 16:9 com faixas pretas nas laterais. Para encher a tela e manter o overlay:

1. No NVIDIA App (**Sistema → Telas**) ou no Painel de Controle NVIDIA (**Tela → Ajustar tamanho e posição da área de trabalho**):
   - Modo de dimensionamento: **Tela cheia**;
   - Executar dimensionamento em: **GPU**.
2. No Windows, em **Configurações → Sistema → Tela**, mude a resolução para **1920 × 1080**.
3. Deixe o SC2 em **"Tela cheia (em janela)"**. O jogo enche a tela e o overlay fica por cima.
4. Depois de jogar, volte o Windows para a resolução nativa (ex.: 2560 × 1080).

### Formato da build

Cada build é um arquivo `.yml` (ou `.yaml`) na pasta de builds. As **chaves são em inglês**. Os **textos podem estar em qualquer língua** e aparecem exatamente como foram escritos, com acentos.

```yaml
name: Terran Bio + Cyclone + Medivac
race: Terran
vs: Zerg                  # opcional
steps:
  - {time: "0:35", action: Supply Depot}
  - {time: "0:48", action: Refinaria, note: "3 SCVs no gás quando terminar"}
  - {time: "1:06", action: Barrack}
reminders:
  - {text: "TRABALHADOR", every_seconds: 12, until: "7:00"}
```

| Chave | Obrigatória | Descrição |
| --- | --- | --- |
| `name` | não | Nome no menu e no overlay. Se vazio, usa o nome do arquivo. |
| `race` | sim | Sua raça: `Terran`, `Protoss`, `Zerg` ou `Random`. |
| `vs` | não | Raça do adversário para a qual a build serve. Usada na sugestão automática. |
| `steps[].time` | sim | Tempo no formato `m:ss`. |
| `steps[].action` | sim | O que fazer. |
| `steps[].note` | não | Observação em fonte menor, embaixo do passo. |
| `reminders[].text` | sim | Texto do lembrete. |
| `reminders[].every_seconds` | sim | Intervalo em segundos (maior que 0). |
| `reminders[].until` | sim | Até quando repetir (`m:ss`). |

- **Ordem:** os passos são ordenados pelo `time` ao carregar. Passos com o mesmo tempo aparecem juntos.
- **Chaves extras:** são ignoradas.
- **Arquivo inválido:** não derruba o programa. Ele aparece no menu como "arquivo.yml (erro)", desabilitado, e o motivo vai para o `overlay.log`.

### Configuração (`%APPDATA%\SC2BuildOverlay\config.yml`)

```yaml
builds_folder: C:\Users\voce\AppData\Roaming\SC2BuildOverlay\builds
selected_build: terran-bio-cyclone-medivac.yml
hotkeys:
  toggle_visible: Ctrl+Alt+H
  toggle_click_through: Ctrl+Alt+J
opacity: 0.7          # opacidade do fundo (0 a 1)
window_position: {x: 20, y: 200}
warning_seconds: 5    # aviso antes de cada passo
sound: false          # bipe curto quando começa o aviso
show_replays: false   # true = acompanha replays também
player_name: ""       # seu nome no SC2, para a sugestão por raça
```

- **Mudanças à mão:** as edições feitas no `config.yml` valem ao reabrir o programa.
- **Formato dos atalhos:** `Ctrl+Alt+H`, com os modificadores `Ctrl`, `Alt`, `Shift` e `Win` e as teclas `A`–`Z`, `0`–`9` e `F1`–`F24`.
- **Atalho ocupado:** se outro programa já usa a combinação, o atalho fica desativado e o log diz qual foi.
- **Administrador:** não é preciso rodar como administrador.

### Menu da bandeja

Clique no ícone perto do relógio do Windows. Ele pode estar escondido na setinha **^**; arraste-o para a barra se quiser que fique sempre visível.

1. **Build order:** lista as builds da pasta. A build em uso fica marcada. Escolher outra troca na hora, mesmo no meio da partida.
2. **Escolher pasta das builds…:** abre o seletor de pasta do Windows.
3. **Simulação:** só no modo `-sim`. Tem pausar, velocidade 1x/2x/4x e nova partida.
4. **Versão x.y.z**
5. **Sair:** fecha tudo e salva a posição da janela, se ela mudou.

---

## Desenvolvimento e testes

### Requisitos

- **Windows 10 ou 11, 64 bits.** O desenvolvimento e os testes foram feitos no Windows. O projeto usa `golang.org/x/sys/windows` e não compila para macOS nem Linux.
- **Go 1.25 ou mais novo.** Foi desenvolvido com Go 1.27.
- **Compilador C só para o detector de corridas (`-race`), que exige CGO:** por exemplo o WinLibs (MinGW-w64):

  ```powershell
  winget install -e --id BrechtSanders.WinLibs.POSIX.UCRT
  ```

  O programa em si **não** usa CGO.

### Comandos

```powershell
go run .                         # roda o overlay (lê o jogo em localhost:6119)
go run . -sim                    # roda com a API simulada, sem o jogo
go run . -sim -sim-speed 4       # simulação 4x mais rápida
go run . -sim -sim-length 6m     # partidas simuladas de 6 minutos
go run . -version                # mostra a versão

go test ./...                    # todos os testes
go test -race ./...              # testes com detector de corridas (precisa de gcc no PATH)
go test ./internal/clock/ -v     # testes de um pacote
go vet ./...                     # análise estática
go generate ./assets             # recria o ícone (.ico/.png) e o recurso do .exe (.syso)
```

### Estrutura

| Pacote | Responsabilidade |
| --- | --- |
| `main.go` | Só junta tudo: flags, config, poller, bandeja, atalhos e janela. |
| `internal/api` | Cliente da SC2 Client API, detecção de início e fim de partida, poller. |
| `internal/clock` | Relógio interpolado: pausa, velocidade, nova partida. |
| `internal/build` | Carregador das builds `.yml`, estados dos passos, lembretes, catálogo da pasta, sugestão por raça, recarga. |
| `internal/overlay` | Estado compartilhado, modelo de exibição (testável sem janela) e a janela Ebitengine. |
| `internal/tray` | Ícone e menu da bandeja (`fyne.io/systray`) e seletor de pasta (`zenity`). |
| `internal/hotkeys` | Parser e registro dos atalhos globais. |
| `internal/config` | `config.yml`, primeira execução, log. |
| `internal/sim` | Servidor falso da API para o modo `-sim`. |
| `assets` | Ícone e build de exemplo embutidos (`embed`). |
| `cmd/genicon` | Gera o ícone. |

**Testes com dados reais:** alguns testes usam respostas reais da API do jogo, gravadas durante o desenvolvimento, em `internal/api/testdata` e `internal/clock/testdata`. Isso inclui uma partida com pausa.

**Documentação do desenvolvimento:** a especificação, o design, as tarefas e o relatório de verificação estão em `.specs/features/overlay/`.

---

## Gerar o build para Windows

### Com o script (recomendado)

No PowerShell, na pasta do projeto:

```powershell
./build.ps1 1.0.1
```

O script:

1. recria os recursos (`go generate ./assets`);
2. roda `go vet`;
3. roda os testes (com `-race` se houver `gcc` no PATH);
4. compila **`bin\overlay.exe`**.

### Manualmente

```powershell
$env:CGO_ENABLED = "0"; $env:GOOS = "windows"; $env:GOARCH = "amd64"
go build -trimpath -ldflags "-H windowsgui -s -w -X main.version=1.0.1" -o bin\overlay.exe .
```

- `CGO_ENABLED=0`: Go puro, sem compilador C.
- `-H windowsgui`: o programa abre sem janela de console.
- `-X main.version=...`: define a versão mostrada no menu e no `-version`. Sem ela, aparece `dev`.
- `-s -w`: remove informações de depuração, o que deixa o arquivo menor.

O resultado é **um único `.exe`** (~18 MB), com fonte e ícone embutidos. Não precisa instalar nada: basta copiar o `overlay.exe` para qualquer pasta e abrir.

---

## Limitações conhecidas

- **Plataforma:** só Windows. **Não testado no macOS** (nem no Linux).
- **Tela cheia exclusiva:** o overlay não aparece nela (veja "Modos de tela do SC2").
- **O que você construiu:** o overlay não sabe. Ele segue só o relógio.
- **Relógio:** a API informa o tempo em segundos inteiros, e o relógio é interpolado. Nos testes com o jogo real, ele bateu com o relógio da partida.
- **Pausa:** ao pausar o jogo, o overlay leva de 1,6 s a 2 s para parar e pode "voltar" um pouco.
- **Replays:** são ignorados por padrão (`show_replays: false`). Não foram testados com o jogo real.
