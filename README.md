# 🧩 Yoink — Git‑Native Secret Manager

> **“GitOps‑style secrets, locally encrypted, globally invisible.”**

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![SOPS](https://img.shields.io/badge/SOPS-Mozilla-blue.svg)
![Age](https://img.shields.io/badge/Age-Encryption-green.svg)
![GitHub](https://img.shields.io/badge/GitHub‑Backend-Enabled-purple.svg)

---

## ⚡ Overview

**Yoink** is a lightweight, **Git‑native secret manager** that lets you manage encrypted secrets directly inside your Git repositories — **securely, offline, and without introducing infrastructure**.

Under the hood, it uses **[SOPS](https://github.com/mozilla/sops)** + **[Age](https://age-encryption.org/)** for cryptographic security — meaning secrets are encrypted locally before Git ever sees them.  
Your vault lives in **GitHub**, but the plaintext never does.

> **Security through simplicity:** developer‑first encryption, no servers, no SaaS, no extra moving parts.

---

## 🔒 Security Positioning

Yoink’s philosophy is **“local‑first security.”**  
Secrets are **encrypted on your machine**, versioned safely in Git, and only decrypted by those whose keys are explicitly authorized in your project’s `.sops.yaml`.

That gives Yoink:

- ✅ **Strong encryption** (Age + SOPS)
- ✅ **Full local ownership** of keys and data
- ✅ **Offline access** with no backend dependencies
- ⚠️ **Flat access model by default** (everyone in `.sops.yaml` can read everything)
- ⚠️ **Manual key rotation and auditing** for now

It’s a **personal‑ or small‑team‑grade security model**, perfect for developers who want to keep encryption strong without managing infrastructure like HashiCorp Vault, AWS Secrets Manager, or Doppler.

---

## 🎯 Who Yoink Is For

| Ideal For                           | Why                                                                                   |
| ----------------------------------- | ------------------------------------------------------------------------------------- |
| Solo developers & indie hackers     | Keep secrets secure across machines and repos without relying on third‑party storage. |
| Small teams (2‑10 developers)       | Manage shared secrets in GitHub privately, using each member’s Age key.               |
| Tight environments or side projects | No cloud signup, no dependency, instant GitOps‑compatible workflow.                   |
| Offline or air‑gapped environments  | All encryption/decryption happens locally; works entirely without internet access.    |

---

## 🚫 When Yoink Is _Not_ the Right Tool

| Situation                                             | Use something else                                             |
| ----------------------------------------------------- | -------------------------------------------------------------- |
| You need strong role‑based access control             | Use **AWS Secrets Manager**, **Vault**, or **Infisical**.      |
| Enterprise audit + compliance required                | Use **Vault**, **Azure Key Vault**, or **GCP Secret Manager**. |
| Many different services, user tiers, or auto‑rotation | Cloud secret managers or centralized SaaS tools.               |
| You can’t trust developers with their own keys        | Managed IAM systems.                                           |

In short:

> Yoink is **not** a bank vault. It’s the **safe under your desk** — secure, controlled, and 100% yours.

---

## ✨ Highlights & Current Features

Everything below is **working** in the current build.

### 🔐 Core Vault Management

- Encrypted secrets managed by **SOPS + Age**
- Per‑project vault initialized with `yoink vault-init`
- GitHub repository automatically used as secure backend
- Project configuration stored in `.yoink.yaml`
- Built‑in **audit**, **debug**, and **status** commands

### ⚡ Developer Flow

- **Fast HTTPS mode** (no full git clone for reads)
- **Quiet git ops by default**, verbose only when needed
- **`--dry-run` mode** on most commands
- **Portable env exports** (`.env`, JSON)
- **Never exposes plaintext** — safe by default

### 🔁 Key Management

- **`yoink key-sync`** — manage private key backups through GitHub:
  - `setup`, `push`, `pull` subcommands
  - Auto‑creates `username/yoink-keys` private repo
  - Light XOR+Base64 obfuscation for now (non‑cryptographic)
- Detects missing keys, verifies restoration, validates repo access

### 🧠 Diagnostics & Visibility

- **`yoink status`** validates dependencies (SOPS, Age, GitHub access)
- **`yoink audit`** shows commit and PR history for the vault
- **`yoink debug`** prints environment and repo state

---

## 🧠 Security Model — In Simple Terms

| Property           | Protection                                                   |
| ------------------ | ------------------------------------------------------------ |
| **At rest**        | AES‑256 encryption via Age; only decrypted locally           |
| **In transit**     | Git + HTTPS; ciphertext only                                 |
| **Access control** | Age private key ownership and repository permissions         |
| **Auditability**   | Git commits and pull requests serve as audit log             |
| **Blast radius**   | If private key is leaked, all secrets tied to it are exposed |

---

## 💬 When to Use Yoink vs Others

| Tool Type                     | Use When                                                | Example Tools                            |
| ----------------------------- | ------------------------------------------------------- | ---------------------------------------- |
| **Local / Git‑native**        | You want full control, no infra, and personal ownership | _Yoink_, _SOPS_, _git‑crypt_             |
| **Developer SaaS**            | You want team dashboards, access control, cloud sync    | _Infisical_, _Doppler_                   |
| **Cloud / Enterprise Vaults** | You need automated rotations, compliance, audit trails  | _AWS/GCP/Azure Secrets Manager_, _Vault_ |

In short:

- **Yoink =** encrypted Git for humans.
- **SaaS managers =** convenience and control, but vendor lock‑in.
- **Enterprise vaults =** compliance and automation, but at high complexity.

---

## 🛠️ Planned Improvements & Roadmap

### 🔒 Security & Key Management

- **Key Lock / Unlock** — integrate with system keychain or GitHub auth session
- **Key Rotation** — rotate Age keys and re‑encrypt vault automatically
- **Improved Key Backup** — encrypt key backups with user’s SSH key (replace XOR)
- **Key Expiry Detection** — audit key age and report stale keys
- **Password‑protected Age keys** — optional local passphrase mode

### 🗂️ Vault Structure & Access Control

- **Environments & Folders** (`dev/`, `staging/`, `prod/`)
- **Teams & Groups** — group Age keys for scoped access
- **Per‑Scope Access Control** — enforce decryption only for authorized keys
- **Multi‑scope SOPS configs** — `.sops.dev.yaml`, `.sops.prod.yaml`, etc.
- **Environment templates** — prebuilt directory layout via `yoink env-init`

### 🔏 Encryption Options

- **GPG fallback** support for enterprise users
- **Dual encryption (Age + GPG)** for mixed environments
- **Config‑driven encryption modes** (`yoink vault-init --gpg`)

### 🧽 Developer UX & Runtime Safety

- **Environment hygiene** — wipe env vars and temp files after `yoink run`
- **Audit diffing** between commits (masked value comparison)
- **Vault integrity verification** (`yoink verify` for corruption checks)
- **Improved JSON output schemas** for scripting and CI parsing
- **Automated GitHub Actions support** for secure decrypt in CI

### 🧠 Policy & Verification

- **Policy linter** detects insecure `.sops.yaml` setups
- **Vault integrity digests** for tamper detection
- **Secret usage scanner** to catch plaintext secrets in repo

### 🧩 UI & Quality of Life

- **Bubbletea TUI** — simple interactive vault browser
- **Local diff/sync helper** — compare local vs remote secrets quickly
- **More commands:** `yoink rotate`, `yoink verify`, `yoink group`

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

| Command                                                    | Description                                  |
| ---------------------------------------------------------- | -------------------------------------------- |
| `yoink init`                                               | Initialize global configuration              |
| `yoink vault-init`                                         | Initialize per‑project vault                 |
| `yoink set <key> <value>`                                  | Add or update a secret (creates PR)          |
| `yoink get <key>`                                          | Retrieve and decrypt a secret                |
| `yoink list`                                               | List all secret keys                         |
| `yoink export`                                             | Export secrets to `.env` or JSON             |
| `yoink run -- <cmd>`                                       | Run a process with injected secrets          |
| `yoink audit`                                              | Show commit and PR history                   |
| `yoink status`                                             | Run health checks and dependency diagnostics |
| `yoink key-sync`                                           | Backup / restore / setup Age keys            |
| `yoink onboard` / `remove-user`                            | Manage user access keys                      |
| `yoink debug`                                              | Debug vault internals                        |
| _(upcoming)_ `yoink rotate`, `yoink group`, `yoink verify` | Key / team / policy extensions               |

---

## 🧩 Example Developer Flow

```bash
# Initialize your global setup
yoink init

# Create a new vault for your project
yoink vault-init

# Add a secret (creates a PR)
yoink set DATABASE_URL postgres://user:pass@db.example.com

# Retrieve or export secrets locally
yoink get DATABASE_URL
yoink export --env-file .env

# Run with secrets injected
yoink run -- npm run dev

# Backup your key
yoink key-sync setup
yoink key-sync push
```

---

## 🛡️ Security Summary

| Category            | Yoink Protects By                                            |
| ------------------- | ------------------------------------------------------------ |
| **Confidentiality** | Local encryption (Age/SOPS)                                  |
| **Availability**    | Git-based versioning for all encrypted secrets               |
| **Integrity**       | Git commit history + optional integrity checks               |
| **Access**          | Developer-held keys only — no third-party backend            |
| **Risk**            | Key loss or compromise = data loss; mitigatable via key-sync |

---

## 🧠 Inspiration

Inspired by  
➡️ [authetoan/gitops-secret-manager-bridge](https://github.com/authetoan/gitops-secret-manager-bridge)  
➡️ Mozilla SOPS + Age model  
➡️ The idea that secrets can live **in plain sight** — encrypted, versioned, and safe.

---

## 🪶 License

MIT © 2025 — built with curiosity, coffee, & good vibes.

---

**Note:**  
Yoink is a secure proof‑of‑concept intended to explore **local‑first, GitOps‑style secret management**.  
It’s safe for personal and small‑team use but not yet enterprise‑compliant.  
Treat it as a **field‑test vault for developers who value simplicity and ownership**.

---
