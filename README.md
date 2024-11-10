# IvyCLI

IvyCLI is a command-line interface for interacting with OpenAI's GPT models. It allows you to send prompts and receive responses directly from your terminal, with support for conversation history, syntax highlighting, and customizable response colors.

## Features

- **Interact with OpenAI's GPT models**: Use GPT-4 or any other specified model.
- **Conversation history**: Optionally save and load encrypted conversation history.
- **Syntax highlighting**: Display responses with syntax highlighting for code blocks.
- **Customizable response color**: Set the color of the assistant's responses using hex color codes.
- **Configurable via JSON file**: Easily adjust settings in a configuration file.

## Installation

### Build from Source

1. **Clone the repository**:

   ```bash
   git clone https://github.com/yourusername/IvyCLI.git
   ```

2. **Install Dependencies**
   
   ```bash
    go mod tidy
    ```

3. **Build the binary**:

   ```bash
   go build -o IvyCLI ./cmd
   ```

4. **Move the binary to a directory in your PATH**:

   ```bash
   sudo mv IvyCLI /usr/local/bin/
   ```

## Configuration

1. **Create a configuration file** (e.g., `config.json`):

   ```json
   {
       "model": "gpt-4o-turbo",
       "system_prompt": "You are a technical assistant. Provide concise, accurate answers to technical questions, assuming the user has a strong background in technology. Focus on brevity and clarity.",
       "response_color": "#1E90FF"
   }
   ```

2. **Move the configuration file to a config path**:

   Create the configuration directory if it doesn't exist:

   ```bash
   mkdir -p ~/.config/ivycli
   ```

   Move the config file:

   ```bash
   mv config.json ~/.config/ivycli/
   ```

## Usage

1. **Set environment variables**:

   Add the API key and config path to your shell profile (`~/.bashrc` or `~/.zshrc`):

   ```bash
   # For Bash
   echo 'export OPENAI_API_KEY=your_openai_api_key' >> ~/.bashrc
   echo 'export IVYCLI_CONFIG_PATH=~/.config/ivycli/config.json' >> ~/.bashrc
   source ~/.bashrc

   # For Zsh
   echo 'export OPENAI_API_KEY=your_openai_api_key' >> ~/.zshrc
   echo 'export IVYCLI_CONFIG_PATH=~/.config/ivycli/config.json' >> ~/.zshrc
   source ~/.zshrc
   ```
   Replace `your_openai_api_key` with your actual API key. 
   
2. **Run IvyCLI**:

   ```bash
   IvyCLI "Your prompt here"
   ```

   **Example**:

   ```bash
   IvyCLI "Explain the difference between concurrency and parallelism."
   ```

3. **Use conversation history** (optional):

   Include the `--history` flag to enable conversation history:

   ```bash
   IvyCLI --history "Your prompt here"
   ```

4. **Disable syntax highlighting** (optional):

   Use the `--no-color` flag to disable syntax highlighting and colored output:

   ```bash
   IvyCLI --no-color "Your prompt here"
   ```

## Encryption of Conversation History

IvyCLI uses encryption to securely store your conversation history. When you enable the `--history` flag, your conversations are saved to a file encrypted using AES-256-GCM. This ensures that sensitive information in your conversations remains confidential.

**How it works**:

- **Passphrase**: You are prompted to enter a passphrase when the application needs to encrypt or decrypt the conversation history.
- **Key Derivation**: The passphrase is used to derive an encryption key using PBKDF2 with SHA-256.
- **Encryption**: The conversation history is encrypted with AES-256-GCM before being saved to disk.
- **Decryption**: When loading the conversation history, you are prompted for the passphrase to decrypt the data.

**Why encryption was added**:

- **Security**: To protect sensitive information that may be part of your conversation history.
- **Privacy**: Ensures that only you can access your saved conversations by requiring a passphrase.

## Best Practices

- **Binary Location**: Place the `IvyCLI` binary in a directory included in your `PATH` environment variable, such as `/usr/local/bin/`, for easy access.
- **Configuration File**: Store your `config.json` in the standard configuration directory `~/.config/ivycli/`.
- **Environment Variables**: Set the necessary environment variables in your shell profile (e.g., `~/.bashrc` or `~/.zshrc`) for persistence.

  ```bash
  # Add these lines to your shell profile
  export OPENAI_API_KEY=your_openai_api_key
  export IVYCLI_CONFIG_PATH=~/.config/ivycli/config.json
  ```

- **Passphrase Management**: Use a strong, unique passphrase for encrypting your conversation history and keep it secure.