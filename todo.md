# 🚀 PROJECT GLOCK: THE MODULAR BLOCK TERMINAL

## 🛠️ PROGRESS TRACKER

### 🟢 PHASE 1: The PTY Engine & Block Foundation
- [x] **Session Domain:** `internal/sh/domain/session.go` [COMPLETED]
- [x] **PTY Port:** `internal/sh/ports/shell_manager.go` [COMPLETED]
- [x] **PTY Adapter:** `internal/sh/adapters/pty_proxy.go` (using `creack/pty`) [COMPLETED]
- [x] **ANSI Parser Logic:** `internal/sh/logic/prompt_detector.go` [COMPLETED]
- [x] **TUI Initialization:** `internal/ui/adapters/bubble_tea_app.go` [COMPLETED]
- [x] **Block Entity:** `internal/ui/domain/block.go` [COMPLETED]
- [x] **Main Entry Point:** `cmd/glock/main.go` [COMPLETED]

### 🟡 PHASE 2: Interactive Smart Input
- [x] **Syntax Port:** `internal/syntax/ports/highlighter.go` [COMPLETED]
- [x] **Tree-sitter Adapter:** `internal/syntax/adapters/ts_bash.go` [COMPLETED]
- [x] **Input Component:** `internal/ui/components/input_field.go` [COMPLETED]
- [x] **Ghost Text Logic:** `internal/ui/logic/prediction_overlay.go` [COMPLETED]

### 🟠 PHASE 3: Contextual Autocompletion & History
- [ ] **History Store:** `internal/completion/adapters/sqlite_store.go` [PENDING]
- [ ] **FS Scanner:** `internal/completion/adapters/fs_provider.go` [PENDING]
- [ ] **Suggester Engine:** `internal/completion/ports/suggester.go` [PENDING]
- [ ] **UI Menu:** `internal/ui/components/suggestion_list.go` [PENDING]

### 🔴 PHASE 4: AI & Final Polishing
- [ ] **AI Port:** `internal/completion/ports/ai_client.go` [PENDING]
- [ ] **OpenAI/Local Adapter:** `internal/completion/adapters/llm_client.go` [PENDING]
- [ ] **Clipboard Port:** `internal/ui/ports/clipboard.go` [PENDING]
- [ ] **Keybinding Manager:** `internal/ui/logic/hotkeys.go` [PENDING]
