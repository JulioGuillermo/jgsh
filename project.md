# 🚀 PROJECT GLOCK: THE MODULAR BLOCK TERMINAL

**Vision:** A TUI-based Terminal Wrapper written in Go that intercepts shell sessions to organize output into interactive "Blocks," providing IDE-grade syntax highlighting and contextual autocompletion.

---

## 🏗️ 1. ARCHITECTURAL REQUIREMENTS (The "Rules of the Game")

### A. Modular Hexagonal Architecture
Each module must have:
* `internal/<module>/domain`: Pure business logic and entities (No dependencies).
* `internal/<module>/ports`: Interfaces (Inbound/Outbound).
* `internal/<module>/adapters`: Implementation (Infrastructure/UI).

### B. Structural Constraints
* **One Struct/Interface per File:** Named after the entity (e.g., `session.go` for `struct Session`).
* **No Global State:** Everything must be injected via constructors (e.g., `NewService(db Port)`).
* **Decoupled Functions:** Logic that doesn't belong to a struct must be in a file named by its functional group (e.g., `ansi_parser.go`).
* **Testing:** Every `port` must have a corresponding `mock` for unit testing the `domain`.
* **Flattened Code:** Maximum 3 levels of nesting. Use guard clauses (`if err != nil { return }`) to avoid `else` blocks.

---

## 🗂️ 2. FOLDER STRUCTURE

```text
.
├── cmd/glock/main.go           # Entry point
├── internal/
│   ├── sh/                     # Shell & PTY Management
│   ├── ui/                     # Bubble Tea TUI & Components
│   ├── syntax/                 # Tree-sitter Highlighting
│   ├── completion/             # History, FS, & AI Suggestions
│   └── shared/                 # Common interfaces/types (minimal)
├── pkg/                        # Exportable utilities (ANSI, etc.)
└── go.mod
```

---

## 🛠️ 3. PROJECT ROADMAP (The Phases)

### 🟢 PHASE 1: The PTY Engine & Block Foundation
**Goal:** Run a shell (Zsh/Bash) inside Go and capture output into discrete data structures.

- [ ] **Define Session Domain:** `internal/sh/domain/session.go` (Tracks PID, state, and current buffer).
- [ ] **PTY Port:** `internal/sh/ports/shell_manager.go` (Interface with `Start()`, `Write()`, `Read()`).
- [ ] **PTY Adapter:** `internal/sh/adapters/pty_proxy.go` using `creack/pty`.
- [ ] **ANSI Parser Logic:** `internal/sh/logic/prompt_detector.go`. Use state machines to detect the return of the Shell Prompt (this signals the end of a "Block").
- [ ] **TUI Initialization:** `internal/ui/adapters/bubble_tea_app.go` with a vertical list of blocks.
- [ ] **Block Entity:** `internal/ui/domain/block.go` (Fields: `ID`, `Command`, `Output`, `ExitCode`, `Timestamp`).

### 🟡 PHASE 2: Interactive Smart Input
**Goal:** Create a command line that feels like an IDE.

- [ ] **Syntax Port:** `internal/syntax/ports/highlighter.go`.
- [ ] **Tree-sitter Adapter:** `internal/syntax/adapters/ts_bash.go`. Map nodes like `command`, `argument`, `string` to `lipgloss.Style`.
- [ ] **Input Component:** `internal/ui/components/input_field.go`. Must support real-time styling of the string while the user types.
- [ ] **Ghost Text Logic:** `internal/ui/logic/prediction_overlay.go` to render suggested text in a dimmed color.

### 🟠 PHASE 3: Contextual Autocompletion & History
**Goal:** Implement the "Knowledge" layer.

- [ ] **History Store:** `internal/completion/adapters/sqlite_store.go`. Save every executed command with its working directory.
- [ ] **FS Scanner:** `internal/completion/adapters/fs_provider.go`. Scans `.` for files/folders.
- [ ] **Suggester Engine:** `internal/completion/ports/suggester.go`. A service that aggregates results from History, FS, and eventually AI.
- [ ] **UI Menu:** `internal/ui/components/suggestion_list.go`. A floating popup that appears above the cursor.

### 🔴 PHASE 4: AI & Final Polishing
**Goal:** Connect to LLMs and finalize UX.

- [ ] **AI Port:** `internal/completion/ports/ai_client.go`.
- [ ] **OpenAI/Local Adapter:** `internal/completion/adapters/llm_client.go`. Sends the last 5 blocks as context for "fix my command" features.
- [ ] **Clipboard Port:** `internal/ui/ports/clipboard.go` to allow copying specific block outputs.
- [ ] **Keybinding Manager:** `internal/ui/logic/hotkeys.go` (Standardized shortcuts like `Ctrl+R`, `Cmd+K`).

---

## 📐 4. MODULE SPECIFICATIONS (For the AI Builder)

### Module: `sh` (Shell)
* **Behavior:** Must use a non-blocking `io.Reader` loop. Every chunk of data read from the PTY must be sent to the UI via a Go `Channel`.
* **Prompt Detection:** To identify "Blocks," look for the sequence `\x1b]133;A\x1b\\` (FTCS_Prompt) or use a regex to detect typical shell prompt endings if the shell doesn't support semantic marks.

### Module: `ui` (User Interface)
* **Renderer:** Use `charmbracelet/lipgloss` for all borders, paddings, and colors.
* **State Management:** The main Bubble Tea model should contain a `[]Block`. When a command starts, a new `Block` is appended. While running, the `Block.Output` is updated.
* **Concurrency:** Use `tea.Tick` and a subscription to the Shell Channel to update the UI without lag.

### Module: `syntax` (The Parser)
* **Granularity:** Do not highlight the whole buffer. Only highlight the active `input_field`.
* **Performance:** Use a debouncer. Re-parse with Tree-sitter only after 10ms of typing inactivity.

---

## 🧪 5. TESTING STRATEGY
* **Unit Tests:** Every domain logic file must have a `_test.go` file (e.g., `history_logic_test.go`).
* **Mocking:** Use `uber-go/mock` or manual structs to simulate the Shell PTY for the UI tests.
* **Isolation:** The `ui` module should be testable without ever actually opening a real shell.

---

## 📋 6. FINAL RECAP FOR GENERATION
1.  **Language:** Go 1.25+.
2.  **Architecture:** Hexagonal / Ports & Adapters.
3.  **UI Framework:** Bubble Tea (Charm).
4.  **Backend:** PTY (`creack/pty`).
5.  **Parsing:** Tree-sitter.
6.  **Style:** SOLID, No nesting, One file per Struct/Interface.

