# Yoink

> **Yoink — a Git‑native, zero‑infrastructure secret manager for developers and teams.**  
> Secure secrets live alongside your code — encrypted, versioned, and shared through GitHub.

---

### 📦 Current Release Status

| Area                                            | Status                                |
| ----------------------------------------------- | ------------------------------------- |
| Init & setup (`init`, `vault-init`)             | ✅ Working                            |
| Manage secrets (`set`, `get`, `delete`, `list`) | ✅ Working                            |
| Runtime injection (`run`)                       | ✅ Working                            |
| Vault workspace & cleanup                       | ✅ Working                            |
| Export secrets (`export`)                       | ⚠️ Bug: causes crash, fix in progress |
| Hide Git logs                                   | 🔜 Pending                            |
| Fast `curl`‑based reads                         | 🔜 Planned                            |
| Multi‑device key sync                           | 🔜 Planned                            |
| GitHub Actions secrets sync                     | 🔜 Planned                            |
| `yoink audit` — vault history + pending PRs     | 🔜 Planned                            |
| TUI / Web UI                                    | 🧪 Research phase                     |
| Cloud Yoink (GitHub App integration)            | 🧩 Long‑term roadmap                  |

---

## 🧠 What Yoink Is

Yoink is a **Git‑backed secret manager** that uses GitHub as its encrypted vault — but hides all the Git.

It combines:

- [Mozilla SOPS](https://github.com/mozilla/sops): Encryption engine
- [Age](https://age-encryption.org/): Key management
- [GitHub](https://github.com/): Distribution, access control & audit trail

No backend, no sync conflicts, no manual merges — just automatic, reliable secret operations.

---

## 🔧 Requirements

Install dependencies:

```bash
brew install sops age gh git
gh auth login
```

Yoink requires:

- `sops` – encryption/decryption
- `age` – key generation and decryption
- `gh` – GitHub CLI for pull requests
- `git` – for committing updates (hidden from the user)

---

## 🚀 Quick Start (Current Functionality)

### 1. Initialize Yoink Globally

```bash
go install github.com/jack-kitto/yoink@latest
yoink init
```

Creates:

- `~/.config/yoink/config.yaml`
- `~/.config/yoink/age.{key,pub}`

---

### 2. Initialize a Project Vault

```bash
cd my-project
yoink vault-init
```

Creates a dedicated GitHub repository (e.g., `my-project-vault`),
sets up `.yoink.yaml`, `.sops.yaml`, and `.gitignore`.

This vault securely holds encrypted secrets.

---

### 3. Add and Retrieve Secrets

```bash
yoink set API_KEY "super-secret"
yoink get API_KEY
yoink list
```

Each `set` automatically:

- Encrypts your value with Age & SOPS
- Commits the change
- Creates a GitHub Pull Request (PR)
- Cleans up the temp clone

“Fetch” commands (`get`, `list`) always read the latest truth directly from the vault — no local workspace to sync.

---

### 4. Inject Secrets at Runtime

```bash
yoink run -- ./deploy.sh
yoink run -- docker-compose up
yoink run -- npm start
```

`yoink run` decrypts secrets temporarily, exports them as environment variables, executes your command, and cleans up.  
Vault clones reside briefly under `~/.config/yoink/vaults/` and are deleted automatically after use.

---

### 5. Export Secrets (⚠️ Known bug)

```bash
yoink export --env-file .env
yoink export --json > secrets.json
```

Currently causes a nil pointer panic (to be fixed in next release).

---

### 6. Collaborate with Team Members

**Grant access:**

```bash
yoink onboard
```

Creates a PR adding your public key to the vault’s `.sops.yaml`.

**Revoke access:**

```bash
yoink remove-user age1abc...
```

Creates a PR removing a user’s access key.

Once merged, everyone with access can run:

```bash
yoink get SECRET
yoink run -- npm start
```

---

## 🧭 Design Philosophy

**State‑less by design.**

Yoink deliberately avoids the ideas of “local vs remote,” “sync,” or “diff.”  
Every command executes directly against the latest source of truth in the vault.

- No working copies
- No manual merges
- No pull/push patterns
- No sync drift

You always interact with **the current, decryptable state**.

---

## ⚙️ Current Limitations

| Issue                                    | Description                                                      |
| ---------------------------------------- | ---------------------------------------------------------------- |
| ❗ **Export crash**                      | `yoink export` causes nil store reference                        |
| 💬 **Verbose Git logs**                  | Raw `git` output still visible; will be hidden by default        |
| 🐢 **Slow read commands**                | Reads clone full vault; will use `curl` to fetch encrypted files |
| 🔐 **No local key sync**                 | You must manually move your Age key between devices              |
| 🧮 **No GitHub Actions integration yet** | Workflows must be created manually                               |
| 🪶 **CLI only (no GUI)**                 | TUI & web UIs are planned for later                              |
| 🧩 **GitHub-only scope**                 | Other Git providers pending future support                       |

---

## 🧩 Future & Roadmap

| Priority                             | Feature                                                                   | Description |
| ------------------------------------ | ------------------------------------------------------------------------- | ----------- |
| 🧩 **Silent Git Mode**               | Hide all git/gh logs unless `--verbose`                                   |
| ⚡ **Fast Reads via HTTPS (`curl`)** | Fetch encrypted file via GitHub Raw or API instead of cloning             |
| 🧠 **`yoink status`**                | Validate setup (config presence, key decryption test, environment health) |
| 📜 **`yoink audit`**                 | Show vault change history **and** pending Pull Requests                   |
| 🔑 **Multi‑Device Key Sync**         | Back up your Age key in a private repo (`@user/yoink-keys`)               |
| ⚙️ **Fix `export` Bug**              | Properly initialize and fetch from fresh vault                            |
| 🧰 **Improved UX & Logging**         | Subtle colors, spacing, clean line output                                 |
| ⚙️ **GitHub Actions Integration**    | Auto‑generate workflow to sync secrets to GitHub Environments             |
| 💾 **Caching for Speed**             | Ephemeral decrypted cache for instant repeat gets                         |
| 🖥️ **TUI Interface**                 | ncurses‑style interactive CLI dashboard                                   |
| 🌐 **Web GUI / WebSocket API**       | Serve Yoink locally via `yoink ui` or `yoink serve`                       |
| ☁️ **Yoink Cloud (Optional)**        | GitHub App integration for centralized secret sync                        |
| 🔁 **Third‑Party Secret Sync**       | Sync to GitHub Envs, AWS Secrets Manager, Doppler, etc.                   |

---

## 📜 `yoink audit`

**Goal:** Transparency without exposing internal Git mechanics.

The `audit` command will display both _historical updates_ and any _pending Pull Requests_ related to secrets.

Example:

```bash
yoink audit
```

Output:

```
Vault: jack-kitto/test-yoink-project-vault

🔐 Recent Updates:
• 2025-11-26  update secret PROD_API_KEY by @jack
• 2025-11-25  delete OLD_TOKEN          by @ci-bot
• 2025-11-24  add SENDGRID_KEY          by @sarah

📬 Pending Pull Requests:
• #17  Update DB_PASSWORD  by @sarah
• #18  Rotate REDIS_URL    by @jack
```

Optional flags:

```bash
--json       # output machine-readable data
--limit 10   # show limited history
--short      # omit pull requests
```

Internally:

- Uses the `gh api` or `gh pr list` commands (authenticated)
- Pulls recent commits and messages
- Never requires local checkout

---

## 🧮 Stateless Command Model

| Command                         | Type         | Purpose                                                     |
| ------------------------------- | ------------ | ----------------------------------------------------------- |
| `yoink init`                    | setup        | Initialize user config and keys                             |
| `yoink vault-init`              | setup        | Prepare new project vault                                   |
| `yoink set` / `yoink delete`    | write        | Add, update, or remove secrets (PR created)                 |
| `yoink get` / `yoink list`      | read         | Fetch decrypted secrets directly from latest vault snapshot |
| `yoink run`                     | read         | Run commands with secrets injected in env                   |
| `yoink export`                  | read         | Export decrypted output to .env or JSON                     |
| `yoink status`                  | diagnostic   | Check key/config validity                                   |
| `yoink audit`                   | transparency | Display history + pending changes                           |
| `yoink onboard` / `remove-user` | access       | Manage team membership in vault                             |
| `yoink doctor`                  | diagnostic   | Verify dependencies (git, gh, sops, age)                    |

All other Git-related behavior is completely hidden.

---

## ⚡ Performance Roadmap

| Level | Optimization              | Expected Gain                       |
| ----- | ------------------------- | ----------------------------------- |
| 1     | Silent Git (default)      | Clean UX                            |
| 2     | HTTPS raw fetch           | ~100× faster read ops               |
| 3     | Shallow clones for writes | Much faster PR generation           |
| 4     | Ephemeral cache           | Instant repeat gets                 |
| 5     | Background agent          | Near-zero latency across CLI/TUI/UI |

---

## 🔑 Multi‑Device Key Sync (Planned)

To securely use Yoink across multiple machines:

```bash
yoink key-sync setup  # creates @<user>/yoink-keys private repository
yoink key-sync pull   # clones and installs Age key on new machine
```

Keys stay under your GitHub account, encrypted and privately stored.

---

## 🧰 Developer Platform Vision

| Interface                   | Purpose                                              |
| --------------------------- | ---------------------------------------------------- |
| **CLI**                     | Primary developer tool (current)                     |
| **TUI (`yoink ui`)**        | Interactive console interface for browsing secrets   |
| **Web GUI (`yoink serve`)** | Local secure web dashboard for encryption/decryption |
| **Cloud (optional)**        | GitHub App service for team‑level vault handling     |
| **GitHub Actions**          | Automated secret propagation for CI/CD               |

Example planned integration:

```bash
yoink actions generate
```

Creates `.github/workflows/yoink-sync.yml`, which:

- Installs Yoink in CI runner
- Decrypts vault secrets securely
- Updates GitHub environment secrets

---

## 🧱 Development

```bash
git clone https://github.com/jack-kitto/yoink.git
cd yoink
make dev     # development build with race detection
make build   # production build in ./bin/yoink
make test
```

---

## 🧑💻 Contributing

We welcome PRs and discussion — especially for:

- Fixing `export` panic
- Implementing `yoink audit`
- Curl‑based read optimizations
- TUI or Web GUI experiments
- GitHub Actions sync prototype
- Key sync design

See [CONTRIBUTING.md](CONTRIBUTING.md) for style and contribution guide.

---

## 🪶 License

MIT — see [LICENSE](LICENSE)

---

**Built by developers who wanted secret management to be invisible, safe, and Git‑native — without feeling like Git.**
