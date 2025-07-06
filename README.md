# Alec CLI

A directory navigator and shell script executor with a beautiful TUI interface.

## Features

- **Directory Navigation**: Browse directories with an intuitive interface
- **Shell Script Execution**: Execute `.sh`, `.bash`, and `.zsh` scripts directly
- **Breadcrumb Navigation**: See your current path and navigate back easily
- **Configurable Root Directory**: Set custom root directory via config file
- **Beautiful TUI**: Built with Charm libraries for a polished experience

## Installation

```bash
go build -o alec
```

## Usage

### Configuration

Before running Alec, you need to create a configuration file at `~/.alec.json`:

```json
{
  "root_dir": "/path/to/your/scripts"
}
```

Example configuration:

```json
{
  "root_dir": "/Users/yourusername/scripts"
}
```

### Basic Usage

Navigate directories and execute scripts:

```bash
./alec
```

If no configuration file exists, Alec will display an error message with instructions on how to create one.

### Controls

- **↑/↓** or **j/k**: Navigate up/down in the list
- **Enter**: Select item (enter directory or execute script)
- **Backspace** or **h**: Go back to parent directory
- **q**: Quit the application

### Directory Structure

The application will show:
- 📁 **Directories** (green) - Navigate into them
- 🚀 **Shell Scripts** (red) - Execute them
- Hidden files (starting with `.`) are ignored

## Configuration File

Alec requires a configuration file at `~/.alec.json` to specify the root directory for navigation. The configuration file uses JSON format and supports the following options:

- `root_dir`: The directory path to use as the root for navigation

Example:
```json
{
  "root_dir": "/Users/yourusername/development/scripts"
}
```

## Development

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework

### Building

```bash
go build -o alec
```

### Running

```bash
go run .
```

### Project Structure

- `main.go` - CLI entry point and Cobra setup
- `model.go` - Bubble Tea model and TUI logic
- `config.go` - Configuration management
- `filesystem.go` - Directory reading and file type detection
- `executor.go` - Shell script execution