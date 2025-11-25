# Yoink MVP

## 🧱 **CLI Core**

These are the commands and behaviours that make Yoink usable from the terminal.

| Category                 | Feature                                                                                                |
| ------------------------ | ------------------------------------------------------------------------------------------------------ |
| **Setup**                | `yoink init` — create global configuration (~/.config/yoink/config.yaml)                               |
|                          | `yoink vault-init` — create `.yoink.yaml` linking the project to its vault repository                  |
| **Secret CRUD**          | `yoink set KEY VALUE` — add or update an encrypted secret                                              |
|                          | `yoink get KEY` — decrypt and print a secret                                                           |
|                          | `yoink list` — list available secret keys                                                              |
| **Secret Injection**     | `yoink run -- <command>` — run any command with decrypted secrets as environment variables             |
| **Exporting**            | `yoink export --env-file <path>` — write decrypted secrets to a `.env` file                            |
|                          | `yoink export --json` — output secrets as JSON                                                         |
| **User Management**      | `yoink onboard` — generate or reuse local Age keys and automatically open a PR adding them to the team |
|                          | `yoink remove-user <email>` _(optional for MVP+1)_ — remove a user and re‑encrypt secrets              |
| **Short Help / Version** | Global help text and `yoink --version` flag                                                            |

---

## 🔐 **Secrets Storage & Encryption**

How secrets are stored, encrypted, and synced.

| Category                  | Feature                                                                                 |
| ------------------------- | --------------------------------------------------------------------------------------- |
| **Storage Format**        | Encrypted files (`*.enc.yaml`) using Mozilla **SOPS**                                   |
| **Encryption Keys**       | Developers use local **Age** key pairs                                                  |
|                           | CI/CD (optional) uses KMS (AWS/GCP/Azure) keys                                          |
| **Automatic Encryption**  | All `set`, `get`, `list`, `export` calls transparently encrypt/decrypt via the SOPS CLI |
| **.sops.yaml management** | Generated automatically with developer public keys and project KMS info                 |
| **Validation**            | Refuse to run if `sops` is not installed                                                |

---

## 🧩 **Per‑Project Vaults**

Separation of secrets for different services.

| Category                     | Feature                                                                       |
| ---------------------------- | ----------------------------------------------------------------------------- |
| **Project Config**           | `.yoink.yaml` created in each project repo                                    |
|                              | Contains path or Git URL of vault repository and default secret file location |
| **Vault Repo Bootstrapping** | Detect org/project name via Git                                               |
|                              | Automatically create vault repo using `gh repo create` if not found           |
|                              | Initialize repository with `.sops.yaml` and readme                            |
| **Isolation**                | Each vault repo encrypts and stores only its project's secrets                |

---

## 💾 **Git/GitHub Integration**

Everything needed to use Git as the transport layer.

| Category                   | Feature                                                                                |
| -------------------------- | -------------------------------------------------------------------------------------- |
| **Auto‑Commit / Push**     | After every `yoink set`, encrypt, commit, and push secrets to the vault repo           |
| **User Onboarding via PR** | `yoink onboard` uses GitHub CLI (`gh`) to create Pull Request adding developer Age key |
| **Offboarding**            | Command or manual PR removing keys from `.sops.yaml`; follow with auto re‑encryption   |

---

## ⚙️ **Configuration**

Persistent settings at both global and per‑project level.

| Scope           | File                          | Purpose                                     |
| --------------- | ----------------------------- | ------------------------------------------- |
| **Global**      | `~/.config/yoink/config.yaml` | Stores default paths and global CLI options |
| **Per Project** | `.yoink.yaml`                 | Defines vault repo URL and file paths       |
| **Per Vault**   | `.sops.yaml`                  | Defines encryption rules and recipient keys |

---

## 🧠 **Developer Experience**

Usability details & guardrails.

| Category                 | Feature                                                                                           |
| ------------------------ | ------------------------------------------------------------------------------------------------- |
| **Dependency Checks**    | Verify that `sops`, `age`, and optionally `gh`, are installed; friendly error messages if missing |
| **Pretty Output**        | Clear emojis / success/failure markers                                                            |
| **Dry‑Run Mode**         | `--dry-run` to simulate without writing to disk or pushing                                        |
| **Git Safety**           | Add `.gitignore` entries for plaintext exported files                                             |
| **Minimal Dependencies** | Single static binary (`go build`)                                                                 |
| **Cross Platform**       | Works on macOS, Linux, and Windows (WSL)                                                          |

---

## 🧰 **Internal Package Responsibilities**

| Package            | Responsibility                                                   |
| ------------------ | ---------------------------------------------------------------- |
| `internal/config`  | Handles global config file load/save                             |
| `internal/project` | Manages `.yoink.yaml`, detects vault repo                        |
| `internal/store`   | Abstracts secret storage, wraps SOPS encryption                  |
| `internal/git`     | Runs Git / GitHub operations (commit, push, pull, repo creation) |
| `cmd/*`            | Implements individual CLI commands                               |

---

## ✅ **Summary of MVP Deliverables**

| Tier                     | Deliverable                                                     |
| ------------------------ | --------------------------------------------------------------- |
| **Core CLI**             | All basic commands compile and run locally                      |
| **Encryption**           | SOPS + Age used automatically for all secret operations         |
| **Vault Management**     | `.yoink.yaml` + GitHub repo creation and syncing                |
| **Developer Workflow**   | Onboard command + environment injection                         |
| **Exports & Automation** | `.env`, JSON, or runtime injection                              |
| **Security Baseline**    | Encrypted at rest, no plaintext in Git, versioned and auditable |
| **Documentation**        | README explaining design and usage                              |
