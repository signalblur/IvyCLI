# IvyCLI

IvyCLI is a command-line tool for interacting with OpenAI's GPT models directly from your terminal. It supports conversation history, syntax highlighting, and customizable response colors.

## Features

- Interact with OpenAI's GPT models (e.g., GPT-4).
- Encrypted conversation history.
- Syntax highlighting for code blocks in responses.
- Customizable response colors.
- Configurable via a JSON file.

## Installation

### Prerequisites

- Go (version 1.16 or newer).

### Clone the Repository

```bash
git clone https://github.com/yourusername/IvyCLI.git
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
go build -o IvyCLI
```

### Move the Binary to Your PATH

Move the binary to a directory included in your `PATH`, such as `/usr/local/bin/`:

```bash
sudo mv IvyCLI /usr/local/bin/
```

## Configuration

### Create a Configuration File

Create a `config.json` file in the project directory:

```json
{
    "model": "gpt-4",
    "system_prompt": "You are a technical assistant. Provide concise, accurate answers to technical questions.",
    "response_color": "#1E90FF",
    "max_history_size": 10
}
```

### Move the Configuration File

Move the `config.json` file to the standard configuration directory:

```bash
mkdir -p ~/.config/ivycli
mv config.json ~/.config/ivycli/
```

## Usage

### Set Environment Variables

Add the required environment variables:

- **OpenAI API Key**:

  ```bash
  export OPENAI_API_KEY="your_openai_api_key"
  ```

- **Passphrase for Encryption**:

  ```bash
  export IVYCLI_PASSPHRASE="your_secure_passphrase"
  ```

- **Configuration File Path**:

  ```bash
  export IVYCLI_CONFIG_PATH="$HOME/.config/ivycli/config.json"
  ```

Consider adding these to your shell profile (`~/.bashrc` or `~/.zshrc`) for persistence:

```bash
# Add these lines to your shell profile
echo 'export OPENAI_API_KEY="your_openai_api_key"' >> ~/.bashrc
echo 'export IVYCLI_PASSPHRASE="your_secure_passphrase"' >> ~/.bashrc
echo 'export IVYCLI_CONFIG_PATH="$HOME/.config/ivycli/config.json"' >> ~/.bashrc
source ~/.bashrc
```

Replace `"your_openai_api_key"` and `"your_secure_passphrase"` with your actual API key and passphrase.

### Run IvyCLI

```bash
IvyCLI "Your prompt here"
```

### Example

```bash
IvyCLI "Explain the difference between concurrency and parallelism."
```

### Reset Conversation History

To reset the conversation history:

```bash
IvyCLI --reset-history
```

### Disable Conversation History

To run without using the conversation history:

```bash
IvyCLI --no-history "Your prompt here"
```

## Encryption of Conversation History

IvyCLI encrypts your conversation history using AES-256-GCM with a passphrase-derived key. The passphrase is provided via the `IVYCLI_PASSPHRASE` environment variable.
