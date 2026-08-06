# Gomander

Gomander is a Go process daemonization library based on [Cobra](https://github.com/spf13/cobra), enabling your program to easily support foreground and background daemon modes with complete process lifecycle management.

## Features

- 🚀 **Subcommand Architecture** - Built on Cobra, providing `start`, `stop`, `restart`, `reload`, `status` subcommands
- 🧩 **Custom Commands** - Register Cobra commands with arguments, flags, and nesting through `WithCommands`
- 🔄 **Daemon Mode** - Support `-d` flag to run process in background
- 📁 **PID File Management** - Automatic creation and cleanup of PID files
- 📝 **Log Redirection** - Automatically redirect output to log file in daemon mode
- 🛑 **Graceful Shutdown** - Support SIGTERM and SIGINT signals for graceful stopping
- ♻️ **Hot Reload** - Support SIGHUP signal to trigger configuration reload
- ⚙️ **Flexible Configuration** - Use functional options pattern to customize PID and log file paths

## Installation

```bash
go get github.com/muleiwu/gomander
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/muleiwu/gomander"
)

func main() {
    gomander.Run(func() {
        fmt.Println("Application starting...")
        
        for {
            time.Sleep(5 * time.Second)
            fmt.Println("Running...")
        }
    })
}
```

### Custom Configuration

```go
func main() {
    gomander.Run(func() {
        // Your business logic
    }, 
        gomander.WithPidFile("./myapp.pid"),
        gomander.WithLogFile("./myapp.log"),
    )
}
```

### Register Custom Commands

Use `WithCommands` to register one or more top-level Cobra commands. Custom commands retain Cobra's argument validation, flags, aliases, and nested subcommands:

```go
import "github.com/spf13/cobra"

func main() {
    versionCmd := &cobra.Command{
        Use:   "version",
        Short: "Show the application version",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("myapp v1.0.0")
        },
    }

    gomander.Run(func() {
        // Your application logic
    }, gomander.WithCommands(versionCmd))
}
```

Calling `WithCommands` more than once accumulates commands. Custom commands execute independently and do not automatically participate in daemon, PID, stop, or reload lifecycle management. Command names and aliases must not conflict with `start`, `stop`, `restart`, `reload`, `status`, `help`, or another registered command. Invalid registrations print an error and exit with status 1.

## Command Line Usage

After compiling your program, you can use the following subcommands:

```bash
go build -o myapp
```

### start - Start Process

```bash
# Run in foreground (logs output to terminal)
./myapp start

# Run as background daemon
./myapp start -d
# or
./myapp start --daemon
```

In daemon mode:
- Process runs in background, detached from terminal
- Logs redirected to log file (default `./runtime/logs/gomander.log`)
- PID saved to file (default `./runtime/gomander.pid`)

### stop - Stop Process

```bash
./myapp stop
```

Reads the PID file and sends SIGTERM signal to gracefully stop the daemon process.

### restart - Restart Process

```bash
./myapp restart
```

Stops the currently running process, then restarts it in daemon mode.

### reload - Reload Configuration

```bash
./myapp reload
```

Sends SIGHUP signal to the daemon process, can be used to trigger configuration reload (requires reload logic implementation in business code).

### status - Check Status

```bash
./myapp status
```

Displays the current status of the daemon process, including:
- Running state (running / stopped)
- Process PID
- PID file path
- Log file path

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithPidFile(path)` | PID file path | `./runtime/gomander.pid` |
| `WithLogFile(path)` | Log file path | `./runtime/logs/gomander.log` |
| `WithCommands(commands...)` | Register additional top-level Cobra commands | None |

## How It Works

### Foreground Mode (start)

```
myapp start → Execute user function directly → Logs output to terminal
```

### Daemon Mode (start -d)

```
myapp start -d → Fork child process → Parent process exits
                      ↓
                Child process (daemon)
                      ↓
                Create new session (setsid)
                      ↓
                Write PID file
                      ↓
                Redirect output to log file
                      ↓
                Execute user function
```

### Signal Handling

| Signal | Behavior |
|--------|----------|
| SIGTERM | Graceful exit, cleanup PID file |
| SIGINT | Graceful exit, cleanup PID file |
| SIGHUP | Trigger reload (does not exit process) |

## Complete Example

See [example/main.go](example/main.go) for a complete example.

```bash
cd example
go build -o myapp

# Start daemon
./myapp start -d

# Check status
./myapp status

# View logs
tail -f myapp.log

# Reload configuration
./myapp reload

# Restart process
./myapp restart

# Stop process
./myapp stop
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GOMANDER_DAEMON=1` | Internal use, identifies current process as daemon child process |

## Notes

1. Ensure you have permission to create PID and log files at the specified paths
2. Before stopping process, ensure PID file exists and process is running
3. Signal handling automatically cleans up PID file
4. `restart` command waits for original process to exit (maximum 10 seconds) before starting new process

## Dependencies

- [cobra](https://github.com/spf13/cobra) - Command line framework

## License

MIT
