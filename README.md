<div align="right">

  ![IvyCLI-Logo](https://github.com/user-attachments/assets/96d078c7-d841-43a7-ba51-9f6f653228c2)

</div>

<div align="center">

# IvyCLI

IvyCLI is a command-line tool for interacting with OpenAI's GPT models directly from your terminal. It supports encrypted conversation history, markdown formatting for responses, and a REPL mode for interactive use.

![Static Badge](https://img.shields.io/badge/mission:-a_simple_and_secure_way_to_interact_with_chatgpt_via_the_cli-purple)

![GitHub top language](https://img.shields.io/github/languages/top/signalblur/ivycli)
![GitHub last commit](https://img.shields.io/github/last-commit/signalblur/ivycli)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

</div>

<img width="1512" alt="Screenshot 2024-11-11 at 12 40 35 AM" src="https://github.com/user-attachments/assets/04bcf1ef-2281-461a-a7b9-33e61f39054e">

## Features

- Interact with OpenAI's GPT models (e.g., GPT-4).
- Encrypted conversation history using AES-256-GCM.
- Markdown formatting for responses.
- REPL mode for interactive use.
- Configurable via a JSON file in `~/.config/ivycli/`.

## Installation

### Prerequisites

- Go (version 1.23.3 or newer).

### Clone the Repository

```bash
git clone https://github.com/signalblur/IvyCLI.git
```

### Install Dependencies

Navigate to the project directory and install the required Go packages:

```bash
cd IvyCLI
go mod tidy
```

### Build the Binary

Build the `IvyCLI` binary:

```bash
CGO_ENABLED=0 go build -o IvyCLI ./cmd
```

### Move the Binary to Your PATH

Move the binary to a directory included in your `PATH`, such as `/usr/local/bin/`:

```bash
sudo mv IvyCLI /usr/local/bin/
```

## First-Time Setup

On the first run, if the configuration directory `~/.config/ivycli/` does not exist, IvyCLI will guide you through a setup process:

- It will prompt you for:
  - OpenAI model (default: `gpt-4`).
  - System prompt (optional, with a default prompt provided).
  - Maximum history size (default: 10).
  - Whether to enable markdown formatting (default: yes).
- It will then set the environment variables for the OpenAI API key and encryption passphrase.

## Configuration

The configuration file is located at `~/.config/ivycli/config.json`. Example:

```json
{
    "model": "gpt-4",
    "system_prompt": "You are a technical assistant. Provide concise, accurate answers to technical questions.",
    "max_history_size": 10,
    "enable_markdown": true
}
```

## Usage

### Run IvyCLI

```bash
IvyCLI "Your prompt here"
```

### Example

```bash
IvyCLI "Explain the difference between concurrency and parallelism."
```

### Enter REPL Mode

For interactive conversations, use the REPL mode:

```bash
IvyCLI --repl
```

Press `Ctrl+C` to exit REPL mode.

### Reset Conversation History

To reset the conversation history:

```bash
IvyCLI --reset-history
```

**Note:** If you some how lose the password, run this to have it use whatever your current password is set to as the environment variable.

### Disable Conversation History

To run without using the conversation history:

```bash
IvyCLI --no-history "Your prompt here"
```

### Disable Markdown Formatting

To disable markdown formatting in the terminal response output:

```bash
IvyCLI --disable-markdown "Your prompt here"
```

**Note:** If you set `"enable_markdown": false` in `~/.config/ivycli/config.json` you will not need to avoid having the `--disable-markdown` flag specified in the CLI everytime.

## Encryption of Conversation History

IvyCLI encrypts your conversation history using AES-256-GCM with a passphrase-derived key. The passphrase is provided via the `IVYCLI_PASSPHRASE` environment variable. This ensures your conversation history remains secure.

## Environment Variables

If not already set during first-time setup, you can manually set the required environment variables:

- **OpenAI API Key**:

  ```bash
  export OPENAI_API_KEY="your_openai_api_key"
  ```

- **Passphrase for Encryption**:

  ```bash
  export IVYCLI_PASSPHRASE="your_secure_passphrase"
  ```

Consider adding these to your shell profile (`~/.bashrc` or `~/.zshrc`) for persistence:

```bash
# Add these lines to your shell profile
echo 'export OPENAI_API_KEY="your_openai_api_key"' >> ~/.bashrc
echo 'export IVYCLI_PASSPHRASE="your_secure_passphrase"' >> ~/.bashrc
```

To make the changes active, run:

```bash
source ~/.bashrc
```
