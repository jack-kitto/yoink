# 🧩 Yoink — Git-Native Secret Manager (MVP)

> **“GitOps‑style secrets, locally encrypted, globally invisible.”**

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![SOPS](https://img.shields.io/badge/SOPS-Mozilla-blue.svg)
![Age](https://img.shields.io/badge/Age-Encryption-green.svg)
![GitHub](https://img.shields.io/badge/GitHub‑Backend-Enabled-purple.svg)

---

## ⚡ Overview

**Yoink** is a lightweight, proof‑of‑concept secret manager that uses **SOPS** + **Age** for encryption and **GitHub** as a decentralized backend vault.

It’s a command‑line tool that makes secrets feel _invisible yet always available_ — no manual git merges, no cloud dependence, just instant, versioned, encrypted configuration.

This version is a **vibe‑coded, MVP‑level experiment** designed to explore how a truly Git‑native, pull‑request‑based secret management system could feel in practice.

---

## ✨ Highlights & Current Features

Everything below is **working** in the current build.

### 🔐 Core Vault Management

- Encrypted secrets managed with **SOPS** and **Age**
- Per‑project vault initialized with `yoink vault-init`
- GitHub repository automatically used as secure backend
- Project config stored in `.yoink.yaml`
- Built-in **audit**, **status**, and **debug** commands

### ⚡ Performance & Developer Experience

- **Fast HTTPS fetch mode** for read-only operations (`get`, `list`, `export`, `run`)  
  → Decrypts locally using SOPS without full git clone
- **Quiet Git operations by default** — no verbose logs unless `--verbose` is passed
- **Faster exports** and zero local-state dependencies
- Support for `--dry-run` across most commands

### 🔁 Key Management

- **`yoink key-sync`** — backup and restore your Age key using a **private GitHub repo**:
  - `yoink key-sync setup` → creates a `username/yoink-keys` repository
  - `yoink key-sync push` → encrypts and backs up your key
  - `yoink key-sync pull` → restores your key securely to a new machine
  - Simple XOR+Base64 obfuscation for backup; private repos enforced

### 🧠 Diagnostics & Visibility

- **`yoink status`**: checks all dependencies, Age key, config, vault access, and GitHub auth
- **`yoink audit`**: lists recent vault commits and pull requests with clean formatting
- **`yoink debug`**: shows repo and file information for troubleshooting

### 🚀 Developer UX

- Machine-friendly JSON output mode (`--json`)
- Emoji-based lightweight summaries for clarity
- Consistent help and flag usage through Cobra
- Safe by default — never exposes plaintext secrets

---

## 💡 Architecture

```
┌───────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   Developer   │───▶│  Yoink CLI (SOPS)│───▶│   GitHub Vault   │
│ (local env)   │    │ (Encrypt/Decrypt)│    │ (Encrypted store)│
└───────────────┘    └──────────────────┘    └──────────────────┘
```

---

## 🧰 Commands Overview

| Command                         | Description                                  |
| ------------------------------- | -------------------------------------------- |
| `yoink init`                    | Initialize global configuration              |
| `yoink vault-init`              | Initialize per-project vault                 |
| `yoink set <key> <value>`       | Add or update a secret (creates PR)          |
| `yoink get <key>`               | Retrieve and decrypt a single secret         |
| `yoink list`                    | List all available secret keys               |
| `yoink export`                  | Export secrets as `.env` or JSON             |
| `yoink run -- <cmd>`            | Run arbitrary commands with injected secrets |
| `yoink audit`                   | Show vault history and open PRs              |
| `yoink status`                  | Perform a full diagnostic check              |
| `yoink key-sync`                | Setup, backup, and restore Age keys          |
| `yoink onboard` / `remove-user` | Manage team access keys                      |
| `yoink debug`                   | Inspect vault repo and metadata              |

---

## 🔧 Installation

```bash
go install github.com/jack-kitto/yoink@latest
```

Or run locally during development:

```bash
make build
./yoink version
```

---

## 🧩 Current State — Proof of Concept (MVP)

This project was **vibe-coded** as a minimal, working proof of concept — its purpose is to test and validate the **“invisible Git vault”** idea rather than achieve production-level polish.

### Goals Achieved ✅

- Working CLI for all CRUD operations
- Fast HTTPS mode (no full repo syncs required for reads)
- Key sync / backup working via GitHub private repos
- Stable `.env` and JSON export
- Robust, quiet git subprocess handling

### Known Limitations ⚠️

- Not optimized for large vaults (due to SOPS runtime cost)
- Limited authentication modes (currently GitHub CLI only)
- Key sync obfuscation is **not strong encryption** (safe only for private repos)
- No TUI or web interface yet

---

## 🛣️ Future Exploration

Yoink was inspired by [authetoan/gitops-secret-manager-bridge](https://github.com/authetoan/gitops-secret-manager-bridge), bringing that idea into a lightweight, zero-config developer UX.

Ideas for future exploration:

- 🔧 **Integrate into existing projects** as a drop‑in secrets backend
- ☁️ Support for other Git providers (GitLab, Bitbucket)
- 🧑💻 TUI mode using Bubbletea for local secret browsing
- 🌐 Web dashboard via `yoink serve`
- ⚙️ GitHub Actions workflow generator (`yoink actions generate`)
- 🪄 Optional Age key cloud sync via GitHub App or GPG fallback

---

## 🧪 Example Developer Workflow

```bash
# Initialize once
yoink init && yoink vault-init

# Add a secret (creates PR)
yoink set DATABASE_URL postgres://user:pass@db.example.com

# View or export locally
yoink get DATABASE_URL
yoink export --env-file .env

# Run with injected environment
yoink run -- npm run dev

# Backup keys
yoink key-sync setup
yoink key-sync push
```

---

## 🧠 Inspiration

This experiment takes inspiration from  
**[authetoan/gitops-secret-manager-bridge](https://github.com/authetoan/gitops-secret-manager-bridge)**

That project elegantly merges GitOps discipline with AWS Secrets Manager.  
Yoink is the local‑first, developer‑centric counterpart — exploring what happens when we rely _only_ on GitHub as the coordination layer and SOPS+Age as the encryption model.

---

## 🪶 License

MIT © 2025 — built with curiosity, coffee, & good vibes.

---

**Note:** This is a prototype meant for experimentation — do not use in production vaults yet.  
Think of it as a _field test_ for next‑generation, pull‑request‑driven secrets management.

---
