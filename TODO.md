# TODO & Future Roadmap 📋

This document outlines upcoming features, architectural improvements, and roadmap goals for `jee-anki`.

---

## 🏛️ 1. UPSC Civil Services & Competitive Exams Expansion
Expand the tool beyond STEM / JEE into a universal, high-retention flashcard generator for UPSC CSE, State PSCs, and humanities-heavy exams.

- [ ] **Exam Profile Selector in TUI:**
  - Add an initial prompt: `Select Exam Track: [JEE / STEM] [UPSC Civil Services] [Custom]`
- [ ] **UPSC High-Yield Prompt Engineering:**
  - **Indian Polity & Governance:** Constitutional Articles, Constitutional & Statutory bodies, Amendments, landmark Supreme Court doctrines (Basic Structure, Kesavananda Bharati, Puttaswamy, etc.).
  - **Modern & Ancient History:** Chronologies, Treaties, Governor-Generals / Viceroys, peasant & tribal uprisings, socio-religious reform movements.
  - **Geography:** Mountain passes, river tributaries, national parks, wildlife sanctuaries, biosphere reserves, ocean currents, soil types.
  - **Economy:** Key indices, reports & publishing bodies, RBI monetary policy instruments, schemes and nodal ministries.
  - **Environment & Ecology:** International conventions (UNFCCC, CBD, Ramsar, CITES), IUCN statuses, pollution standards.
- [ ] **Hierarchical Deck Taxonomy for UPSC:**
  - Auto-categorize into `UPSC::GS1`, `UPSC::GS2`, `UPSC::GS3`, `UPSC::Prelims-Facts` inside the `Inbox`.

---

## 🗂️ 2. Core Engine & Workflow Improvements

- [ ] **Multi-PDF Batch Processing:**
  - Allow selecting multiple PDFs or "Process All PDFs in Folder" with a unified progress bar and aggregated card count.
- [ ] **Cloze Deletion Note Type (`{{c1::...}}`):**
  - Add support for generating Anki `Cloze` notes in addition to standard `Basic` (Front/Back) cards.
- [ ] **Diagram & Media Extraction (Anki Media Collection):**
  - Detect high-value figures and diagrams in PDFs (e.g. physics circuits, organic reaction schemes, geography maps).
  - Crop and extract images using `pdfcpu` and upload them directly via AnkiConnect's `storeMediaFile` API.
  - Embed `<img>` tags directly into note fields.
- [ ] **Configurable Target Deck:**
  - Allow users to configure an alternative target deck name in `./jee-anki.config.json` (defaults to `Inbox`).
- [ ] **Smart Rate-Limit Backoff:**
  - Implement exponential backoff with jitter when encountering HTTP 429 errors from Google Gemini API on large PDF batches.

---

## 🖥️ 3. Platform & UI Enhancements

- [ ] **macOS Release Support:**
  - Package `.command` double-clickable launcher or lightweight notarized `.app` wrapper to bypass Gatekeeper quarantine.
- [ ] **Web UI / Optional Local Dashboard:**
  - Add an optional `--web` flag that spawns a lightweight local browser dashboard (`http://localhost:8080`) for users who prefer point-and-click browser forms over terminal windows.
- [ ] **Anki Web & Mobile Sync Trigger:**
  - Add an optional post-sync trigger via AnkiConnect (`sync` action) to immediately push new cards to AnkiWeb.
