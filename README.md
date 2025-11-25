# Yoink

> **Yoink — a Git‑native, no‑backend secret manager for developers and teams.**  
> Pull secrets. Push code. No servers.

---

## 🧠 What Yoink Is

Yoink is a **CLI tool** that manages your project secrets directly from Git — not from a hosted backend.  
It wraps [Mozilla SOPS](https://github.com/mozilla/sops) for encryption, [Age](https://age-encryption.org/) for keys, and uses GitHub as the coordination point for teams.

Each project you work on has its own **vault repository** that holds encrypted secrets.  
Developers run the `yoink` CLI to initialize, edit, inject, or synchronize those secrets.  
CI/CD pipelines can decrypt them automatically using organization keys.

---

## 🔧 Prerequisites

Before using Yoink, make sure you have these tools installed:

```bash
# Install SOPS (encryption)
brew install sops
# or: go install go.mozilla.org/sops/v3/cmd/sops@latest

# Install Age (key management)
brew install age
# or: go install filippo.io/age/cmd/age@latest

# Install GitHub CLI (repository management)
brew install gh
# or: https://cli.github.com/

# Install Git (version control)
brew install git
# or: https://git-scm.com/
```

Authenticate with GitHub:

```bash
gh auth login
```

---

## 🚀 Quick Start Walkthrough

### Step 1: Global Setup

First, initialize Yoink's global configuration:

```bash
# Install yoink
go install github.com/jack-kitto/yoink@latest

# Initialize global config
yoink init
```

This creates `~/.config/yoink/config.yaml` and sets up your local environment.

### Step 2: Project Setup

Navigate to your project directory and initialize a vault:

```bash
cd your-project/
yoink vault-init
```

This will:

- Generate an Age key pair if you don't have one
- Create `.yoink.yaml` linking to a vault repository
- Create the vault repository on GitHub (e.g., `your-project-vault`)
- Set up `.sops.yaml` for encryption
- Add appropriate `.gitignore` entries

### Step 3: Managing Secrets

Store your first secret:

```bash
yoink set API_KEY "your-secret-api-key"
yoink set DB_PASSWORD "super-secret-password"
yoink set REDIS_URL "redis://localhost:6379"
```

Each secret is encrypted with SOPS and automatically committed to your vault repository.

Retrieve secrets:

```bash
# Get a specific secret
yoink get API_KEY

# List all secret keys
yoink list
```

### Step 4: Using Secrets in Development

Inject secrets into your development commands:

```bash
# Run your app with secrets loaded
yoink run -- npm start

# Run Docker Compose with secrets
yoink run -- docker-compose up

# Run any command with secrets
yoink run -- python manage.py runserver
```

Export secrets for external tools:

```bash
# Create .env file for local development
yoink export --env-file .env

# Export as JSON
yoink export --json > secrets.json

# Print to stdout
yoink export
```

### Step 5: Team Collaboration

When a new team member joins:

```bash
# They run this in the project directory
yoink onboard
```

This automatically:

1. Generates their Age key if needed
2. Forks the vault repository
3. Creates a pull request adding their public key
4. Notifies maintainers for approval

When the PR is merged, they can access all project secrets.

To remove a user:

```bash
yoink remove-user age1publickey123...
```

This creates a PR to remove their access from the vault.

---

## 📖 Detailed Walkthrough

### Setting Up Your First Project

Let's walk through setting up Yoink for a new web application:

#### 1. Initialize the Project

```bash
mkdir my-web-app
cd my-web-app
git init
echo "# My Web App" > README.md
git add . && git commit -m "Initial commit"

# Create GitHub repo
gh repo create --public

# Set up Yoink vault
yoink vault-init
```

#### 2. Add Development Secrets

```bash
# Database configuration
yoink set DB_HOST "localhost"
yoink set DB_PORT "5432"
yoink set DB_NAME "myapp_dev"
yoink set DB_USER "developer"
yoink set DB_PASSWORD "dev123"

# API keys for external services
yoink set STRIPE_API_KEY "sk_test_123..."
yoink set SENDGRID_API_KEY "SG.abc123..."

# Check what we have
yoink list
```

#### 3. Use Secrets in Development

Create a simple Node.js app:

```bash
# Initialize Node.js project
npm init -y
npm install express pg

# Create app.js
cat > app.js << 'EOF'
const express = require('express');
const app = express();

const config = {
  db: {
    host: process.env.DB_HOST,
    port: process.env.DB_PORT,
    database: process.env.DB_NAME,
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
  },
  stripe: {
    apiKey: process.env.STRIPE_API_KEY,
  }
};

app.get('/', (req, res) => {
  res.json({
    message: 'Hello World!',
    config: {
      db: { ...config.db, password: '[REDACTED]' },
      stripe: { apiKey: '[REDACTED]' }
    }
  });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
EOF
```

Run with secrets injected:

```bash
# Run the app with secrets automatically loaded
yoink run -- node app.js

# Or export for manual use
yoink export --env-file .env
source .env && node app.js
```

#### 4. Set Up Production Secrets

Add production environment secrets:

```bash
yoink set PROD_DB_HOST "prod-db.company.com"
yoink set PROD_DB_PASSWORD "very-secure-password"
yoink set STRIPE_LIVE_KEY "sk_live_456..."
```

### Working with Teams

#### Adding a New Developer

When Sarah joins your team:

1. **Sarah clones the project:**

   ```bash
   git clone git@github.com:yourorg/my-web-app.git
   cd my-web-app
   ```

2. **Sarah requests vault access:**

   ```bash
   yoink onboard
   ```

3. **You (as maintainer) receive a PR** in the vault repository with Sarah's public key

4. **You review and merge the PR** to grant Sarah access

5. **Sarah can now use secrets:**
   ```bash
   yoink list
   yoink get DB_PASSWORD
   yoink run -- npm start
   ```

#### Rotating a Compromised Secret

If an API key gets compromised:

```bash
# Update the secret
yoink set STRIPE_API_KEY "sk_test_new_key_789..."

# The encrypted file is automatically committed and pushed
# All team members get the update on next git pull
```

#### Removing a Team Member

When someone leaves:

```bash
# Get their public key from the vault's .sops.yaml
yoink remove-user age1ql3z0hjgxjk2k3l1d52j2fv0q6k2g0j1l3k4j5l6m7n8o9p0

# This creates a PR to remove their access and re-encrypt all secrets
```

### CI/CD Integration

#### GitHub Actions Example

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install dependencies
        run: |
          # Install yoink and dependencies
          go install github.com/jack-kitto/yoink@latest
          curl -LO https://github.com/mozilla/sops/releases/download/v3.8.1/sops-v3.8.1.linux
          sudo mv sops-v3.8.1.linux /usr/local/bin/sops
          sudo chmod +x /usr/local/bin/sops

      - name: Set up SOPS key
        run: echo "${{ secrets.SOPS_AGE_KEY }}" > ~/.config/yoink/age.key

      - name: Deploy with secrets
        run: |
          # Secrets are automatically injected
          yoink run -- ./deploy.sh
```

Set up the `SOPS_AGE_KEY` secret in GitHub Actions with a dedicated CI Age key.

### Advanced Usage

#### Dry Run Mode

Test commands without making changes:

```bash
# See what would happen without actually doing it
yoink --dry-run set NEW_SECRET "test-value"
yoink --dry-run onboard
yoink --dry-run remove-user age1abc...
```

#### Environment-Specific Secrets

Organize secrets by environment:

```bash
# Development secrets
yoink set DEV_API_URL "http://localhost:3000"
yoink set DEV_DEBUG_MODE "true"

# Production secrets
yoink set PROD_API_URL "https://api.company.com"
yoink set PROD_DEBUG_MODE "false"

# Use environment-specific exports
yoink export --env-file .env.development
yoink export --env-file .env.production
```

#### Multiple Projects

Each project has its own vault:

```bash
cd project-a/
yoink vault-init  # Creates project-a-vault repository

cd ../project-b/
yoink vault-init  # Creates project-b-vault repository

# Secrets are completely isolated between projects
```

---

## 🎯 Command Reference

| Command                   | Description                          | Example                        |
| ------------------------- | ------------------------------------ | ------------------------------ |
| `yoink init`              | Initialize global configuration      | `yoink init`                   |
| `yoink vault-init`        | Initialize project vault             | `yoink vault-init`             |
| `yoink set <key> <value>` | Store or update a secret             | `yoink set API_KEY abc123`     |
| `yoink get <key>`         | Retrieve a secret value              | `yoink get API_KEY`            |
| `yoink list`              | List all secret keys                 | `yoink list`                   |
| `yoink delete <key>`      | Delete a secret                      | `yoink delete OLD_KEY`         |
| `yoink run -- <command>`  | Run command with secrets injected    | `yoink run -- npm start`       |
| `yoink export`            | Export secrets as .env or JSON       | `yoink export --env-file .env` |
| `yoink onboard`           | Request access to vault (creates PR) | `yoink onboard`                |
| `yoink remove-user <key>` | Remove user access (creates PR)      | `yoink remove-user age1abc...` |
| `yoink version`           | Show version information             | `yoink version`                |

### Global Flags

- `--dry-run`: Show what would be done without making changes
- `--help`: Show help for any command

### Export Options

- `--env-file <path>`: Export as .env format to file
- `--json`: Export as JSON format

---

## ⚙️ How It Works

### 1. Per‑project vaults

Running `yoink vault-init` inside a project creates:

- `.yoink.yaml` configuration file linking to a vault repository
- A GitHub repository (e.g., `yourproject-vault`) containing:
  - Encrypted secrets (`secrets.enc.yaml`) managed by SOPS
  - `.sops.yaml` defining encryption rules with team Age keys
  - Git history providing audit trail

### 2. Encryption Flow

```
Developer adds secret → SOPS encrypts with Age keys → Git commit → GitHub vault repo
                                    ↓
Developer retrieves secret ← SOPS decrypts with local key ← Git pull ← GitHub vault repo
```

### 3. Team Management

- Each developer has an Age key pair (`~/.config/yoink/age.{key,pub}`)
- `yoink onboard` creates PRs to add public keys to vault access
- Vault maintainers approve/deny access via GitHub PR review
- Secrets are re-encrypted when team membership changes

### 4. Security Model

| Component            | Security                                   |
| -------------------- | ------------------------------------------ |
| Local secrets        | Encrypted at rest with SOPS + Age          |
| Vault repository     | Contains only encrypted files              |
| Network transmission | Git over SSH/HTTPS                         |
| Key management       | Age keys, optionally KMS for CI            |
| Access control       | GitHub repository permissions + SOPS rules |

---

## 🧩 Why Use Yoink?

### ✅ Advantages

- **No backend**: All coordination through Git/GitHub
- **Familiar tools**: Uses SOPS, Age, Git, and GitHub
- **Secure**: Encrypted at rest, no plaintext transmission
- **Transparent**: Git history = audit log
- **Portable**: Single binary CLI, no infrastructure
- **Developer‑friendly**: Feels like a local tool, not a service
- **Team‑friendly**: PR‑based access management

### ⚖️ Trade‑offs

- **Requires GitHub**: Currently GitHub-specific (GitLab support planned)
- **Tool dependencies**: Needs SOPS, Age, Git, and GitHub CLI
- **Git‑based**: Audit history tied to Git repository access
- **Key management**: Developers responsible for Age key security

### 🆚 Compared to Alternatives

| Feature             | Yoink | HashiCorp Vault | AWS Secrets Manager | Doppler |
| ------------------- | ----- | --------------- | ------------------- | ------- |
| Backend required    | ❌    | ✅              | ✅                  | ✅      |
| Git‑native          | ✅    | ❌              | ❌                  | ❌      |
| Audit via Git       | ✅    | ⚠️              | ⚠️                  | ⚠️      |
| Team PR workflow    | ✅    | ❌              | ❌                  | ❌      |
| Offline access      | ✅    | ❌              | ❌                  | ❌      |
| Infrastructure cost | $0    | $$$             | $$$                 | $$$     |

---

## 🔧 Configuration Files

### Global Config (`~/.config/yoink/config.yaml`)

```yaml
secrets_file: ~/.config/yoink/secrets.enc.yaml
default_vault: ""
```

### Project Config (`.yoink.yaml`)

```yaml
vault: git@github.com:yourorg/yourproject-vault.git
secrets_file: .yoink/secrets.enc.yaml
```

### SOPS Config (`.sops.yaml` in vault)

```yaml
creation_rules:
  - path_regex: .*\.(yaml|yml|json)$
    age: >-
      age1alice...,
      age1bob...,
      age1charlie...
```

---

## 🛠️ Development

### Building from Source

```bash
git clone https://github.com/jack-kitto/yoink.git
cd yoink
make build

# Binary will be in ./bin/yoink
```

### Running Tests

```bash
make test
```

### Development Build

```bash
make dev  # Builds with race detection
```

---

## 🪶 License

MIT License — see [LICENSE](LICENSE) for details.

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

**Made with 💚 by developers who wanted secret management to feel like `git`.**
