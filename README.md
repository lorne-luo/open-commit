# opencommit

**AI-Powered, Conventional Commit Messages with OpenAI-compatible APIs**

opencommit helps you write clear, conventional, and meaningful Git commit messages automatically using OpenAI-compatible AI providers. Save time, improve your commit history, and focus on what matters—your code.

---

## ✨ Features

- **AI-Generated Commit Messages:** Analyzes staged changes and suggests concise, descriptive messages.
- **AI-Generated Pull Requests:** `opencommit pr` pushes your branch and opens a GitHub PR with an AI-generated title and body.
- **Conventional Commits:** Follows best practices for readability and automation.
- **Multi-Provider Support:** Works with OpenAI, Anthropic, and any OpenAI-compatible endpoint (including local models like Ollama).
- **Customizable Output:** Tune message style, language, and length to fit your workflow.
- **Smart Issue Detection:** Detects and references issue numbers from branch names.
- **Automatic Push:** Push committed changes with `--push`.
- **Cross-Platform:** Linux, macOS, and Windows.

---

## ✅ Requirements

- Go 1.24+ (for `go install`)
- Git
- API key for your chosen AI provider
- GitHub CLI (`gh`) — only for `opencommit pr`

---

## 🚀 Quickstart

```sh
# 1. Install
go install github.com/lorne-luo/open-commit@latest

# 2. Configure your API key
opencommit config set api.key <your-api-key>

# 3. (Optional) Choose a model and/or custom endpoint
opencommit config set api.model gpt-4o
opencommit config set api.baseurl https://your-proxy

# 4. Stage and commit
git add <file>
opencommit
```

---

## 🛠️ Installation

### From Source

```sh
go install github.com/lorne-luo/open-commit@latest
```

Make sure `$HOME/go/bin` is on your `PATH`:

```sh
# Zsh
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc && source ~/.zshrc

# Bash
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc && source ~/.bashrc
```

### Standalone Binary

Download from the [releases page](https://github.com/lorne-luo/open-commit/releases) and place in your `PATH`:

- Linux: `$HOME/.local/bin/` or `/usr/local/bin/`
- macOS: `/usr/local/bin/`
- Windows: `%LocalAppData%\Programs\`

### NixOS

```nix
environment.systemPackages = [ pkgs.opencommit ];
```

---

## ⚙️ Configuration

Get an API key:

- **OpenAI:** https://platform.openai.com/api-keys
- **Anthropic:** https://console.anthropic.com/
- **Local models (Ollama, etc.):** no key needed

```sh
# Set / get / list
opencommit config set api.key <your-api-key>
opencommit config get api.key
opencommit config list
```

Config is stored at `~/.config/opencommit/config.toml` (macOS: `~/Library/Application Support/opencommit/config.toml`).

### Available Keys

```text
[api]
api.key             AI provider API key
api.model           AI model name (default: gpt-3.5-turbo)
api.baseurl         Custom base URL for OpenAI-compatible APIs

[commit]
commit.language     Language for commit messages (default: english)
commit.max_length   Maximum length of commit message (default: 72)

[behavior]
behavior.stage_all    Stage all tracked changes (default: false)
behavior.auto_select  Let AI pick files and message (default: false)
behavior.no_confirm   Skip confirmation prompt (default: false)
behavior.quiet        Suppress output (default: false)
behavior.push         Push to remote after commit (default: false)
behavior.dry_run      Run without making changes (default: false)
behavior.show_diff    Show diff before committing (default: false)
behavior.no_verify    Skip git commit-msg hook (default: false)
```

### Config File Format (TOML)

```toml
[api]
key = "your-api-key"
model = "gpt-3.5-turbo"
baseurl = "https://api.anthropic.com"  # optional
```

### Multi-Provider Examples

**OpenAI** (default endpoint):

```sh
opencommit config set api.key sk-your-openai-key
opencommit config set api.model gpt-4o
```

**Anthropic:**

```sh
opencommit config set api.key sk-ant-your-anthropic-key
opencommit config set api.baseurl https://api.anthropic.com
opencommit config set api.model claude-3-sonnet-20240229
```

**Local Ollama:**

```sh
opencommit config set api.baseurl http://localhost:11434
opencommit config set api.model llama3:8b
# No API key needed
```

---

## 📖 Usage

```sh
git add <file>
opencommit
```

Review the AI-generated message, accept or edit it, and opencommit will create the commit.

### Pull Requests

```sh
opencommit pr              # push and open a PR
opencommit pr --draft      # create as draft
opencommit pr --dry-run    # preview without pushing
```

Combine with `--yes -q`, `--show-diff`, `--language`, `--baseurl`, etc.

### Common Flags

```sh
opencommit --dry-run                 # preview without committing
opencommit --show-diff               # display diff before committing
opencommit --max-length 50           # message length limit
opencommit --language spanish        # generate in another language
opencommit --issue "#123"            # reference an issue
opencommit --no-verify               # skip commit-msg hook
opencommit --push                    # push after commit
opencommit --baseurl https://...     # override endpoint
opencommit --model gpt-4o            # override model
```

### Auto Issue Detection

Issue numbers are detected from branch names:

- `feature-123-description` → `#123`
- `fix-456-bug` → `#456`
- `#789-feature` → `#789`
- `issue-101` → `#101`

### Combining Options

```sh
opencommit --dry-run --show-diff --max-length 60 --language spanish
opencommit --issue "#123" --push --no-verify
opencommit --baseurl https://your-proxy.example.com --model gpt-4o
```

Run `opencommit --help` for the full list.

---

## 🤝 Contributing

Issues and pull requests are welcome.

---

## 📄 License

GPLv3 — see [LICENSE](LICENSE).
