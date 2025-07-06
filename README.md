# Alec CLI

A directory navigator and shell script executor with a beautiful TUI interface.

## Features

- **Directory Navigation**: Browse directories with an intuitive interface
- **Shell Script Execution**: Execute `.sh`, `.bash`, and `.zsh` scripts directly
- **Breadcrumb Navigation**: See your current path and navigate back easily
- **Configurable Root Directory**: Set custom root directory via command line flag
- **Beautiful TUI**: Built with Charm libraries for a polished experience

## Installation

```bash
go build -o alec
```

## Usage

### Basic Usage

Navigate directories and execute scripts:

```bash
./alec
```

### Custom Root Directory

Specify a custom directory to navigate:

```bash
./alec --dir /path/to/your/scripts
./alec -d ~/my-scripts
```

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

## Configuration

By default, Alec uses the directory where the binary is located as the root directory. You can override this with the `--dir` flag.

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