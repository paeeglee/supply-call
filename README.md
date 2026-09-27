# SC2 Build Overlay

Overlay para treinar build orders no StarCraft II. É uma janela pequena, sempre por cima do jogo, que acompanha o relógio da partida e mostra o que fazer agora, o que vem a seguir e o que já passou.

- Só lê a **SC2 Client API oficial** (`http://localhost:6119`), que o próprio jogo abre enquanto está rodando.
- **Não** lê a memória do jogo, **não** injeta nada e **não** envia teclas ou cliques para o jogo.
- É guiado só pelo tempo: você não aperta nada durante a partida.

## Requisitos

- Windows 10 ou 11 (64 bits).
- StarCraft II em **"Tela cheia (em janela)"** (veja abaixo).
- Para compilar: Go 1.25 ou mais novo. Não precisa de compilador C.

## Modo de exibição do SC2 (importante)

O overlay só aparece por cima do jogo em **Tela cheia (em janela)**:

**Opções → Gráficos → Modo de exibição → Tela cheia (em janela)** (em inglês: *Windowed (Fullscreen)*)

| Modo do SC2 | Overlay |
| --- | --- |
| Tela cheia (em janela) | ✅ Aparece. Visualmente igual à tela cheia. |
| Janela | ✅ Aparece, mas o jogo fica numa janela menor que a tela. |
| Tela cheia (exclusiva) | ❌ **Não aparece.** Testado. Desenhar sobre a tela cheia exclusiva exigiria injetar código no jogo, o que este app não faz. |

### Monitor ultrawide (21:9)

O SC2 não suporta ultrawide: em "Tela cheia (em janela)" ele desenha em 16:9 e deixa faixas pretas nas laterais, até no menu. A tela cheia exclusiva enche a tela porque troca a resolução do monitor, mas nela o overlay não aparece.

Para encher a tela **com** o overlay, faça a mesma troca de resolução pelo Windows:

1. **Escala pela GPU:** no NVIDIA App (Sistema → Telas) ou no Painel de Controle NVIDIA (Tela → Ajustar tamanho e posição da área de trabalho), use **Modo de dimensionamento: Tela cheia** e **Executar dimensionamento em: GPU**. Ou ajuste o "Aspect/Proporção" no menu do próprio monitor para **Full**.
2. **Resolução do Windows:** em Configurações → Sistema → Tela → Resolução da tela, escolha **1920 × 1080**.
3. **Modo do SC2:** "Tela cheia (em janela)". O jogo enche a tela esticado, igual à tela cheia exclusiva, e o overlay fica por cima.
4. **Depois de jogar:** volte o Windows para a resolução nativa (ex.: 2560 × 1080).

Testado num monitor 2560×1080.

## Como compilar

Com PowerShell, na pasta do projeto:

```powershell
./build.ps1 1.0.0
```

O script gera os recursos, roda `go vet` e os testes, e compila `bin\overlay.exe`. Ou compile direto:

```powershell
$env:CGO_ENABLED = "0"; $env:GOOS = "windows"; $env:GOARCH = "amd64"
go build -ldflags "-H windowsgui -X main.version=1.0.0" -o bin\overlay.exe .
```

- `-H windowsgui` evita abrir uma janela de console.
- `-X main.version=...` define a versão (sem ela, aparece `dev`).
- A fonte (Go Regular/Bold) e o ícone ficam dentro do `.exe`. O ícone do arquivo vem de `rsrc_windows_amd64.syso`, que já está no repositório. Para gerar de novo: `go generate ./assets`.

## Como rodar

Dê dois cliques em `overlay.exe`, ou:

```powershell
.\bin\overlay.exe            # modo normal: lê o jogo
.\bin\overlay.exe -sim       # simulação, sem o jogo
.\bin\overlay.exe -version   # mostra a versão
```

Na primeira execução, o app cria:

- `%APPDATA%\SC2BuildOverlay\config.yml`: configuração;
- `%APPDATA%\SC2BuildOverlay\builds\terran-bio-cyclone-medivac.yml`: uma build de exemplo, já selecionada;
- `%APPDATA%\SC2BuildOverlay\overlay.log`: log. É zerado ao abrir se passar de 1 MB.

Com o jogo fechado, o overlay mostra "Aguardando partida…". Ao entrar numa partida, o relógio começa sozinho, e a build recomeça a cada partida nova.

### Modo simulação (`-sim`)

Sobe um servidor falso em `localhost:6119` que imita a API: menu (3 s) → carregamento (3 s) → partida de 10:00 → menu, em loop. Serve para testar builds sem abrir o jogo.

```powershell
.\bin\overlay.exe -sim -sim-speed 4          # 4x mais rápido
.\bin\overlay.exe -sim -sim-length 6m        # partidas de 6 minutos
```

No modo `-sim`, a bandeja ganha o submenu **Simulação**, com Pausar, Velocidade 1x/2x/4x e Nova partida. Se o jogo estiver aberto, a porta 6119 fica ocupada e o overlay avisa "Porta 6119 ocupada (jogo aberto?)".

## O que aparece no overlay

- **Topo:** relógio da partida e nome da build. Quando a build acaba, aparece **"Build concluída"**.
- **Linha de lembretes:** por exemplo "TRABALHADOR 0:07", com a contagem até o próximo lembrete. A linha pisca quando o lembrete dispara.
- **Lista com todos os passos:**
  - passos que já passaram: cinza e riscados;
  - passo atual: amarelo, com contagem (ex.: "Supply Depot em 0:05");
  - nos últimos 5 s: a linha pisca em laranja;
  - na hora: "Supply Depot — AGORA" em laranja por 3 s, e depois fica riscado;
  - passos futuros: branco, com o tempo planejado.
- Passos com o mesmo tempo ficam destacados juntos.
- A janela cresce para caber a build inteira. Se a build não couber na tela, a lista rola sozinha para manter o passo atual visível, e os passos que já passaram continuam na lista.

## Formato da build

Cada build é um arquivo `.yml` (ou `.yaml`) na pasta de builds. As **chaves são em inglês**. Os **textos podem estar em qualquer língua** e aparecem exatamente como foram escritos, com acentos.

```yaml
name: Terran Bio + Cyclone + Medivac
race: Terran
vs: Zerg                     # opcional
steps:
  - {time: "0:35", action: Supply Depot}
  - {time: "0:48", action: Refinaria, note: "3 SCVs no gás quando terminar"}
  - {time: "1:06", action: Barrack}
  - {time: "2:51", action: 2º Command Center}
reminders:
  - {text: "TRABALHADOR", every_seconds: 12, until: "7:00"}
```

| Chave | Obrigatória | Descrição |
| --- | --- | --- |
| `name` | não | Nome no menu e no overlay. Se vazio, usa o nome do arquivo. |
| `race` | sim | Sua raça: `Terran`, `Protoss`, `Zerg` ou `Random`. |
| `vs` | não | Raça do adversário para a qual a build serve. Usada na sugestão automática. |
| `steps` | sim | Lista de passos (pelo menos um). |
| `steps[].time` | sim | Tempo no formato `m:ss` (ex.: `1:06`, `10:30`). |
| `steps[].action` | sim | O que fazer. |
| `steps[].note` | não | Observação em fonte menor, embaixo do passo. |
| `reminders` | não | Lembretes periódicos. |
| `reminders[].text` | sim | Texto do lembrete. |
| `reminders[].every_seconds` | sim | Intervalo em segundos (maior que 0). |
| `reminders[].until` | sim | Até quando repetir (`m:ss`). |

- Os passos podem estar fora de ordem: o app ordena pelo `time` ao carregar.
- Chaves extras são ignoradas, então dá para acrescentar campos para outras ferramentas.
- Um arquivo inválido (YAML quebrado, `time` fora do formato, raça errada) não derruba o app. Ele aparece no menu como **"arquivo.yml (erro)"**, desabilitado, e o motivo vai para o `overlay.log`.
- Pode editar a build com o overlay aberto: a pasta é verificada a cada 2 s, e o menu e a build em uso se atualizam sozinhos.

## Menu da bandeja

Clique no ícone perto do relógio do Windows (talvez esteja escondido na setinha **^**):

1. **Build order:** lista as builds da pasta. A build em uso fica marcada. Escolher outra troca na hora (até no meio da partida, usando o tempo atual) e salva a escolha. Se a pasta estiver vazia, aparece "Nenhuma build encontrada".
2. **Escolher pasta das builds…:** abre o seletor de pasta do Windows. Se você cancelar, nada muda.
3. **Simulação:** só no modo `-sim`.
4. **Versão x.y.z:** informativo.
5. **Sair:** fecha tudo e salva a posição da janela.

## Atalhos globais

| Atalho padrão | Ação |
| --- | --- |
| `Ctrl+Alt+H` | Esconder/mostrar o overlay. |
| `Ctrl+Alt+J` | Ligar/desligar o **modo mover**: desliga o click-through. |

- O overlay abre com **click-through ligado**: os cliques passam para o jogo.
- Para mover a janela, aperte `Ctrl+Alt+J`. Aparecem a borda amarela e o texto "modo mover". Arraste com o botão esquerdo; a roda do mouse rola a lista. Aperte `Ctrl+Alt+J` de novo para voltar ao click-through. A posição fica salva.
- Para trocar as teclas, edite `hotkeys` no `config.yml` e reabra o app. Formato: `Ctrl+Alt+H`, com os modificadores `Ctrl`, `Alt`, `Shift` e `Win` e as teclas `A`–`Z`, `0`–`9` e `F1`–`F24`. Evite `F12`: o Windows a reserva.
- Se outro programa já usa a combinação, o atalho fica desativado e o `overlay.log` diz qual foi. Foi o que aconteceu com `Ctrl+Alt+M` no PC de desenvolvimento.
- **Não é preciso rodar como administrador.** Se o SC2 estiver rodando como administrador e os atalhos não funcionarem dentro do jogo, rode o overlay também como administrador.

## config.yml

```yaml
builds_folder: C:\Users\voce\AppData\Roaming\SC2BuildOverlay\builds
selected_build: terran-bio-cyclone-medivac.yml
hotkeys:
  toggle_visible: Ctrl+Alt+H
  toggle_click_through: Ctrl+Alt+J
opacity: 0.7          # opacidade do fundo (0.1 a 1)
window_position: {x: 20, y: 200}
warning_seconds: 5    # aviso antes de cada passo
sound: false          # bipe curto quando começa o aviso
show_replays: false   # true = acompanha replays também
player_name: ""       # seu nome no SC2, para a sugestão por raça
```

Mudanças feitas à mão no `config.yml` valem ao reabrir o app.

### Sugestão automática pela raça do adversário

Se `player_name` estiver preenchido com o seu nome no jogo, o app escolhe sozinho, a cada partida, a primeira build da pasta com `race` igual à sua raça e `vs` igual à raça do adversário. Essa troca não é salva no config.

A escolha manual sempre vence: se você escolher uma build no menu, ela fica até a partida acabar.

## Limitações

- Não funciona sobre **tela cheia exclusiva** (veja acima).
- O overlay não sabe o que você construiu; ele segue só o relógio. A SC2 API não informa isso.
- A API informa o tempo em segundos inteiros e às vezes demora ou falha. O relógio é interpolado e fica a menos de ~0,5 s do jogo. Uma ou duas consultas perdidas não interrompem a partida.
- Ao pausar o jogo, o relógio para 1,6 s a 2 s depois da pausa (pode adiantar um pouco e voltar).
- Replays são ignorados por padrão (`show_replays: false`).
- A janela aparece na barra de tarefas.

## Desenvolvimento

```powershell
go test ./...
go test -race ./...   # precisa de um gcc (ex.: WinLibs) no PATH; o .exe continua sem CGO
```

Pacotes: `internal/api` (cliente da SC2 API e detecção de partida), `internal/clock`, `internal/build`, `internal/overlay`, `internal/tray`, `internal/hotkeys`, `internal/config`, `internal/sim`. O `main.go` só junta tudo. As respostas reais da API usadas nos testes estão em `internal/api/testdata` e `internal/clock/testdata`.
