# Overlay: contexto

**Gathered:** 2026-09-26
**Spec:** `.specs/features/overlay/spec.md`
**Status:** Ready for design

---

## Feature Boundary

Overlay Windows guiado só pelo `displayTime` da SC2 Client API: relógio, lista completa da build com estados passado/agora/atual/futuro, lembretes, bandeja com seleção de build e pasta, dois atalhos globais, modo simulação e um único `.exe`.

---

## Implementation Decisions

### Atalhos

- Padrão: Ctrl+Alt+H (esconder/mostrar) e Ctrl+Alt+J (click-through).
- F12 fica fora: o Windows o reserva para o depurador.

### Passo atual

- Atual = próximo grupo a vencer, com contagem "Ação em m:ss".
- No tempo, o grupo mostra "AGORA" por 3 s e depois fica riscado; o próximo vira atual.

### Visual

- Fundo preto com a opacidade configurada (padrão 0,7), texto branco.
- Atual em amarelo; aviso de 5 s piscando laranja/amarelo; "AGORA" em laranja; passados em cinza riscados.

### Sugestão por raça

- `player_name` no config identifica o jogador; o adversário é o outro.
- A escolha manual feita desde a última partida sempre vence.

### Agent's Discretion

- Layout fino (margens, espaçamentos), formato da linha de lembretes, controles da simulação.

### Declined / Undiscussed Gray Areas → Assumptions

- Estado inicial do click-through, forma de esconder, fonte, lembretes, fim de partida, rolagem manual, controle da simulação e tolerância a jitter: registrados em Assumptions no spec.

---

## Specific References

Mockup escolhido:

```
│ 3:04        Terran Bio...    │
│ SCV!                         │
├──────────────────────────────┤
│ 0:35 Supply Depot  (cinza, riscado)
│▶3:09 Supply Depot em 0:05    │ amarelo/laranja
│ 3:16 Refinaria        (branco)│
```

---

## Deferred Ideas

- (Implementado depois, a pedido: OVL-12, `WS_EX_TOOLWINDOW`.)
