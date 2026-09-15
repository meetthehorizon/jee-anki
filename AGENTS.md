# AGENTS.md 🤖

A concise guide for AI agents and developers working on the `jee-anki` codebase.

---

## 🎯 Project Overview & Core Principles

`jee-anki` is a cross-platform (Linux & Windows) CLI/TUI application written in pure Go that turns study notes PDFs into high-yield, atomic Anki flashcards via Google Gemini and AnkiConnect.

### Non-Negotiable Core Principles:
1. **Zero CGO / Pure Go:** Every dependency must be 100% pure Go (`CGO_ENABLED=0`). Cross-compilation to Windows `.exe` and Linux amd64 must work without external C libraries.
2. **Crash-Proof Terminal Windows:** On Windows `.exe` and Linux desktop double-clicks, the terminal window must **never** close abruptly on success, error, or panic. The deferred `waitForExit()` guard in `cmd/jee-anki/main.go` must remain intact.
3. **Strict Source Grounding:** Cards must strictly reflect the provided document text and formulas. No AI commentary, conversational filler, or invented analogies.
4. **Anki MathJax Standards:** Always format math equations with `\( ... \)` (inline) and `\[ ... \]` (block). **Never** use `$ ... $` or `$$ ... $$` (Anki Desktop does not render raw dollar signs as MathJax by default).
5. **Dynamic Model Targeting:** Use `gemini-flash-latest` as the primary default so the app automatically tracks Google's latest production Flash model without breaking when older model versions are retired.
6. **Zero External Runtime:** Students must never be required to install Python, Poppler, or system libraries.

---

## 🏗️ Architecture & Package Map

```
jee-anki/
├── cmd/
│   └── jee-anki/
│       └── main.go          # Application orchestrator, exit pause, terminal launcher
├── internal/
│   ├── anki/                # AnkiConnect JSON-RPC client (ping, ensureDeck, addNotes)
│   ├── config/              # Local ./jee-anki.config.json store, model & tag history
│   ├── gemini/              # Google Gemini REST client, inline base64 PDF, system prompt & schema
│   ├── pdf/                 # pdfcpu pure-Go page counter and 5-page chunk splitter
│   ├── terminal/            # Linux TTY detector & auto-terminal emulator spawner
│   └── tui/                 # Charm Huh interactive forms, Lipgloss styling, setup guides
├── .github/workflows/
│   └── release.yml          # GitHub Actions cross-platform build & release
├── flake.nix / .envrc       # Hermetic Nix development environment
├── README.md                # Student-facing user guide
└── TODO.md                  # Future roadmap (UPSC expansion, cloze cards, etc.)
```

---

## 🛠️ Development & Testing Workflow

### 1. Environment Setup
```bash
# Uses Nix Flake (contains Go 1.26+, gopls, gotools)
direnv allow
# or:
nix develop
```

### 2. Running Unit Tests
All packages include unit tests with mock HTTP servers for AnkiConnect and Gemini:
```bash
go test -count=1 -v ./...
```

### 3. Cross-Compilation Verification
Verify both builds compile with zero CGO:
```bash
# Linux static binary
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o jee-anki-linux-amd64 ./cmd/jee-anki

# Windows executable
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o jee-anki-windows-amd64.exe ./cmd/jee-anki
```

---

## ⚠️ Important Pitfalls to Avoid

- **Do NOT rely on system-installed PDF tools:** Always use `github.com/pdfcpu/pdfcpu` APIs for PDF operations.
- **Do NOT hardcode local filesystem paths:** Relative paths must resolve against the current working directory (`./jee-anki.config.json`).
- **Do NOT commit API keys or config files:** `./jee-anki.config.json` is explicitly ignored in `.gitignore`.
- **Remember Linux File Manager Semantics:** When launching without a TTY on Linux, `internal/terminal.EnsureTerminal()` spawns a terminal emulator (`alacritty`, `konsole`, `gnome-terminal`, etc.) running the binary with `JEE_ANKI_IN_TERMINAL=1`. Do not break this detection loop.
