# STATE

## Decisions

### AD-001
- **Decision**: O Ebitengine fica com a thread principal; a bandeja roda em goroutine própria com `runtime.LockOSThread()` + `systray.Run`; toda troca de estado com a interface passa por `overlay.Shared` (mutex).
- **Reason**: `systray.RunWithExternalLoop` no Windows roda o loop de mensagens em outra thread (verificado em `systray_windows.go`), e o Ebitengine exige a main thread.
- **Trade-off**: Duas threads de UI; a bandeja não pode ser testada automaticamente.
- **Scope**: `main.go`, `internal/tray`, `internal/overlay`
- **Date**: 2026-09-26
- **Status**: active

### AD-002
- **Decision**: Regras de exibição ficam em funções puras (`build.Timeline`, `build.ReminderStatus`, `overlay.BuildView`); `Draw` só pinta o resultado.
- **Reason**: Testar sem janela e sem o jogo.
- **Trade-off**: Uma camada a mais entre estado e desenho.
- **Scope**: `internal/build`, `internal/overlay`
- **Date**: 2026-09-26
- **Status**: active

## Handoff

