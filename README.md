# git-sage 🧙‍♂️

**git-sage** is a production-ready CLI tool written in Go that leverages AI (OpenRouter) to generate high-quality, emoji-prefixed Git commit messages. It streamlines your workflow by automatically crafting concise subjects and detailed bodies, ensuring your commit history follows [Conventional Commits](https://www.conventionalcommits.org/) best practices.

## ✨ Features

- 🧠 **AI-Powered**: Generates concise, present-tense commit subjects and detailed bodies explaining the _why_.
- 🚀 **Powered by OpenRouter**: Uses the OpenAI-compatible API with configurable models.
- 🎨 **Emoji Support**: Automatically detects commit type and suggests appropriate emojis.
- 🤖 **Context Aware**: Analyzes your staged diff to understand the changes.
- 🕹️ **Interactive TUI**: Keyboard-friendly picker for commit types and scopes (powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)).
- ⚡ **Highly Configurable**: Supports per-repo and global configuration, including branch-based defaults and remote configs.
- 🔒 **Conventional Commits**: Enforces standards while remaining flexible.

## 🚀 Installation

### Prerequisites

- Go 1.21+
- An OpenRouter API Key (Get one at [openrouter.ai](https://openrouter.ai))

### From Source

```bash
git clone https://github.com/Amul-Thantharate/git-sage.git
cd git-sage
make install
```

This will build the binary and move it to `/usr/local/bin/git-sage`. Ensure `/usr/local/bin` is in your `PATH`.

## 🛠 Setup

1. **Set your OpenRouter API Key** (Required):

   ```bash
   export OPENROUTER_API_KEY="your_api_key_here"
   ```

   Add this to your `~/.bashrc`, `~/.zshrc`, or `.env`.

2. **Verify Installation**:
   Since the binary is named `git-sage`, Git automatically recognizes it as a subcommand.
   ```bash
   git sage --help
   ```

## 📖 Usage

1. **Stage your changes**:

   ```bash
   git add .
   ```

2. **Run the tool**:

   ```bash
   git sage
   ```

3. **Interactive Flow**:
   - **Select Type**: Use arrow keys to select the commit type (feat, fix, etc.). It auto-selects a suggestion based on your changes.
   - **Select Scope** (Optional): If configured, select a scope for your commit.
   - **Review & Confirm**: The AI generates a message. Press **Enter** to commit or **n** to abort.

### Flags

- `-message <text>`: Provide a natural language description to guide the AI (e.g., `git sage -message "fix login bug for mobile users"`).
- `--auto`: Skip the interactive picker and confirmation. Uses auto-detected type and generated message immediately.
- `--dry-run`: Print the generated command but do not execute the commit.
- `--no-ai`: Skip AI generation and use a simple fallback message.

## ⚙️ Configuration

Configuration is loaded in the following priority:

1. `.git-sage.yaml` (Repository root)
2. `~/.config/git-sage/config.yaml` (Global config)
3. Remote configuration (if specified in local/global config)

### Example `config.yaml`

```yaml
# Custom emojis for commit types
emojis:
  feat: "✨"
  fix: "🐛"
  docs: "📚"
  custom: "🦄"

# Pre-defined scopes for the interactive picker
scopes:
  - ui
  - api
  - core
  - deps

# Automatically select commit type based on branch name pattern
branch_defaults:
  "feature/": feat
  "fix/": fix
  "docs/": docs

# Sync configuration from a remote URL
remote_config: "https://example.com/company-git-sage.yaml"
```

## 🌐 Environment Variables

| Variable                  | Description                            | Default                        |
| ------------------------- | -------------------------------------- | ------------------------------ |
| `OPENROUTER_API_KEY`      | **(Required)** Your OpenRouter API Key | None                           |
| `OPENROUTER_API_BASE_URL` | Custom endpoint for OpenRouter API     | `https://openrouter.ai/api/v1` |
| `OPENROUTER_MODEL`        | Primary model to use via OpenRouter    | `qwen/qwen3.6-plus:free`       |

If the primary model fails, git-sage automatically retries with fallback model `minimax/minimax-m2.5:free`.

## 🏗 Internal Architecture

- **`internal/git`**: Git command wrappers and diff extraction.
- **`internal/detect`**: Heuristic engine for guessing commit types from file extensions and diffs.
- **`internal/ai`**: OpenRouter API client and prompt engineering.
- **`internal/config`**: Multi-source configuration loader (YAML).
- **`internal/ui`**: TUI implementation using Bubble Tea and Lip Gloss.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**Amul Thantharate** ([@blackroute](https://github.com/blackroute))
