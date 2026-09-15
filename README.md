# jee-anki ⚡

> **Turn your JEE & STEM PDF notes into high-yield Anki flashcards with a single click.**  
> Built for students preparing for JEE Main, JEE Advanced, NEET, and university STEM exams.

---

## 🚀 Quick Start (In 3 Simple Steps)

### Step 1: Install the Anki Add-on (One-Time Setup)
1. Open **Anki Desktop** on your computer.
2. In the top menu, click **Tools** → **Add-ons** → **Get Add-ons...**
3. Paste the AnkiConnect code: **`2055492159`** and click **OK**.
4. *(Recommended for Chemistry)* Also paste the SMILES code: **`1819582297`** (Convert SMILES to SVG).
5. **Restart Anki Desktop**.

### Step 2: Download `jee-anki`
1. Go to the [**Releases Page**](https://github.com/meetthehorizon/jee-anki/releases).
2. Download for your operating system:
   - **Windows:** `jee-anki-windows-amd64.exe`
   - **Linux (Recommended):** `jee-anki-linux-amd64.tar.gz` (extract it into your study folder)
   - **Linux (Standalone Binary):** `jee-anki-linux-amd64`
3. Put the downloaded file in a folder where you keep your study PDFs.
   - *Linux tip:* If you downloaded the standalone binary, make sure it is marked executable (`chmod +x jee-anki-linux-amd64` or right-click → Properties → Permissions → "Allow executing as program").

### Step 3: Run and Extract!
1. Drop your PDF notes (e.g. `Current_Electricity.pdf`, `Calculus_Formulas.pdf`, `Aldehydes_and_Ketones.pdf`) in the **same folder** as `jee-anki`.
2. **Double-click** `jee-anki` (or double-click `run.sh` / `jee-anki.desktop` on Linux) to run it:
   - On Linux, it will automatically detect and open your desktop terminal (Alacritty, Konsole, GNOME Terminal, etc.).
   - It will check that Anki is running.
   - It will ask for your free Google Gemini API Key (stored locally so you only enter it once).
   - Pick your PDF and watch it extract flashcards with live progress!
3. Review and approve the AI-suggested tags (e.g. `physics, current-electricity, circuits`).
4. **Done!** Open Anki Desktop — all new cards are waiting for you in the **`Inbox`** deck.

---

## 🔑 Getting Your 100% Free Google Gemini API Key

You do **not** need a credit card, and the free tier gives you up to **1,500 requests per day**:

1. Visit [**Google AI Studio** (aistudio.google.com)](https://aistudio.google.com/).
2. Sign in with your Google account.
3. Click **"Get API key"** and then **"Create API key"**.
4. Copy your key (starts with `AIzaSy...`).
5. Paste it when `jee-anki` prompts you. It is saved in `./jee-anki.config.json` in your local folder so you never have to re-enter it.

---

## 🧠 What Kind of Cards Does It Create?

- **Atomic & High-Yield:** Instead of overwhelming you with huge walls of text, it breaks complex concepts into bite-sized questions and triggers.
- **MathJax Math Formulas:** Clean LaTeX rendered natively in Anki using `\( ... \)` and `\[ ... \]`. No messy unrendered dollar signs.
- **Organic & Inorganic Chemistry:** Reagents, reaction conditions, and molecular structures rendered with standard SMILES tags (`[smiles]...[/smiles]`) alongside full IUPAC/common names.
- **Physics Circuits & Shortcuts:** Balanced Wheatstone conditions, symmetry shortcuts, time constants (\(\tau = RC\), \(\tau = L/R\)), and resonance conditions.
- **Strict Source Grounding:** Cards only contain facts, definitions, and formulas directly from your notes — no AI fluff, no hallucinations.

---

## 📥 How the "Inbox" Deck Works

To make sure your main study decks stay clean and organized:
1. All new flashcards are placed in a deck named **`Inbox`**.
2. Open Anki and go to **Browse** (or click the `Inbox` deck).
3. Quickly flip through the cards to verify them.
4. Select the cards and move them into your main subject decks (`Physics`, `Chemistry`, or `Maths`).

---

## ❓ Frequently Asked Questions (FAQ)

### Q: It says "Could not connect to Anki Desktop"
- Make sure Anki Desktop is **open and running** on your computer before launching `jee-anki`.
- Ensure you installed the **AnkiConnect** add-on (`2055492159`) and restarted Anki.

### Q: Windows shows "Windows protected your PC" (SmartScreen)
- Because `jee-anki` is a free, open-source tool without an expensive commercial code-signing certificate, Windows SmartScreen may show a blue warning.
- Click **"More info"** and then click **"Run anyway"**.

### Q: Which Gemini model should I choose?
- **`gemini-2.0-flash` (Recommended):** Blazing fast, multimodal PDF vision, and completely free.
- **`gemini-2.5`:** Great if your notes are extremely dense or complex.
- You can also pick **Enter custom model name...** if you want to experiment with newer models.

### Q: How do tags work across multiple sessions?
`jee-anki` automatically remembers the tags you use in `./jee-anki.config.json`. When you parse new notes, it suggests consistent tags (e.g. normalizing `math` to `maths`) so your Anki tag taxonomy stays tidy.

---

## 🛠️ Developer Setup & Building from Source

### Using Nix Flake (Recommended)
This repository includes a hermetic `flake.nix` with Go and developer tools:

```bash
# Allow direnv or enter nix develop
direnv allow
# or:
nix develop

# Run unit tests
go test -v ./...

# Build binary
go build -o jee-anki ./cmd/jee-anki
```

### Standard Go Build
```bash
# Build for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o jee-anki-linux-amd64 ./cmd/jee-anki

# Build for Windows (.exe)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o jee-anki-windows-amd64.exe ./cmd/jee-anki
```
