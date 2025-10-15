[![Support Palestine](https://raw.githubusercontent.com/Safouene1/support-palestine-banner/master/banner-project.svg)](https://github.com/Safouene1/support-palestine-banner)

# opencommit

**AI-Powered, Conventional Commit Messages with OpenAI-compatible APIs**

![Preview](./assets/Screenshot_20241112_103154.png)

**opencommit** helps you write clear, conventional, and meaningful Git commit messages automatically using OpenAI-compatible AI providers. Save time, improve your commit history, and focus on what matters—your code.

---

## ✨ Features

- **AI-Generated Commit Messages:** Let AI analyze your staged changes and suggest concise, descriptive commit messages.
- **AI-Generated Pull Requests:** Use `opencommit pr` to push your branch and open a GitHub pull request with an AI-generated conventional title (and body).
- **Customizable Output:** Tailor the message style and structure to fit your workflow.
- **Conventional Commits:** Ensures messages follow best practices for readability and automation.
- **Cross-Platform:** Works on Linux, Windows, and macOS.
- **Open Source:** Free to use and contribute.
- **Automatic Push:** Push committed changes to remote repository with `--push` flag.
- **Advanced Customization:** Fine-tune commit messages with various flags and options.
- **Smart Issue Detection:** Automatically detects and references issue numbers from branch names.
- **Multi-Provider Support:** Works with OpenAI, Anthropic, and any OpenAI-compatible API endpoints.
- **Custom API Endpoints:** Configure custom base URLs for any AI provider.

---

## 🚀 Quickstart

```sh
# 1. Install (Go required)
go install github.com/lorne-luo/opencommit@latest

# 2. Get your AI API key
#    OpenAI: https://platform.openai.com/api-keys
#    Anthropic: https://console.anthropic.com/

# 3. Configure your API key
opencommit config key set <your-api-key>

# 4. (Optional) Configure model and base URL
opencommit config model set gpt-4o              # optional: change model
opencommit config baseurl set https://your-proxy # optional: custom endpoint

# 5. Stage your changes
git add <file>

# 6. Generate and commit
opencommit
```

---

## 🛠️ Installation

- **From Source:**

  Add To Path:
  - **Zshrc:**

    ```sh
    echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
    source ~/.zshrc
    ```

  - **Bashrc:**

    ```sh
    echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc
    source ~/.bashrc
    ```

  ```sh
  go install github.com/lorne-luo/opencommit@latest
  ```

- **Standalone Binary:**
  Download from the [releases page](https://github.com/lorne-luo/opencommit/releases) and move to a directory in your `PATH`:
  - Linux: `$HOME/.local/bin/` or `/usr/local/bin/`
  - Windows: `%LocalAppData%\Programs\`
  - macOS: `/usr/local/bin/`

- **NixOS:**

  ```nix
  environment.systemPackages = [
    pkgs.opencommit
  ];
  ```

---

## ⚙️ Configuration

### Basic Setup

1. Get your AI API key:
   - **OpenAI:** [OpenAI Platform](https://platform.openai.com/api-keys)
   - **Anthropic:** [Anthropic Console](https://console.anthropic.com/)
   - **Local models:** No API key needed (just set base URL)
2. Set your key:

   ```sh
   opencommit config key set <your-api-key>
   ```

### Advanced Configuration

Configure additional settings using the `opencommit config` command:

```sh
# Set or change the AI model (default: gpt-3.5-turbo)
opencommit config model set gpt-4o
opencommit config model show

# Set custom API base URL (OpenAI, Anthropic, or local endpoints)
opencommit config baseurl set https://api.anthropic.com  # Anthropic
opencommit config baseurl set http://localhost:11434     # Local Ollama
opencommit config baseurl show

# Clear custom base URL (revert to default OpenAI)
opencommit config baseurl set ""

# View current API key
opencommit config key show
```

All configuration is stored in `~/.config/opencommit/config.toml` (on macOS: `~/Library/Application Support/opencommit/config.toml`).

#### Configuration File Format

The configuration file uses TOML format:

```toml
[api]
key = "your-api-key"
model = "gpt-3.5-turbo"
baseurl = "https://api.anthropic.com"  # optional - for Anthropic, Ollama, etc.
```

#### Multi-Provider Examples

**OpenAI:**
```sh
opencommit config key set sk-your-openai-key
opencommit config model set gpt-4o
# No baseurl needed (uses default Openai endpoint)
```

**Anthropic:**
```sh
opencommit config key set sk-ant-your-anthropic-key
opencommit config baseurl set https://api.anthropic.com
opencommit config model set claude-3-sonnet-20240229
```

**Local Ollama:**
```sh
opencommit config baseurl set http://localhost:11434
opencommit config model set llama3:8b
# No API key needed for local models
```

---

## 📖 Usage

1. Stage your changes:

   ```sh
   git add <file>
   ```

2. Run the CLI to generate a commit:

   ```sh
   opencommit
   ```

3. Review and edit the AI-generated message if needed.
4. opencommit will commit your changes with the generated message.

### Create Pull Requests

Use AI to draft a PR title & body and open a GitHub pull request:

```sh
opencommit pr              # opens a ready-for-review PR
opencommit pr --draft      # create as draft
opencommit pr --dry-run    # preview without pushing
```

You can combine `--yes -q`, `--show-diff`, `--language`, `--baseurl`, and other flags just like the commit command.

### Advanced Usage & Customization

#### Commit Message Customization Flags

```sh
# Preview commit without making changes
opencommit --dry-run

# Display the diff before committing
opencommit --show-diff

# Set maximum commit message length (default: 72 characters)
opencommit --max-length 50

# Generate commit messages in different languages
opencommit --language spanish
opencommit --language french

# Reference specific issue numbers
opencommit --issue "#123"
opencommit --issue "JIRA-456"

# Skip git commit-msg hook verification
opencommit --no-verify

# Push committed changes to remote repository
opencommit --push

# Use custom API endpoint
opencommit --baseurl https://your-proxy.example.com

# Use specific AI model
opencommit --model gpt-4o
```

#### Auto Issue Detection

opencommit automatically detects issue numbers from branch names using common patterns:

- `feature-123-description` → references issue #123
- `fix-456-bug` → references issue #456
- `#789-feature` → references issue #789
- `issue-101` → references issue #101

#### Combining Options

```sh
# Comprehensive example: dry run with diff, custom length, and language
opencommit --dry-run --show-diff --max-length 60 --language spanish

# Production workflow: commit and push with issue reference
opencommit --issue "#123" --push --no-verify

# Using custom endpoint with specific model
opencommit --baseurl https://your-proxy.example.com --model gpt-4o
```

For more options:

```sh
opencommit --help
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to open an issue or submit a pull request.

---

## 📄 License

This project is licensed under the GPLv3 License. See the [LICENSE](LICENSE) file for details.
