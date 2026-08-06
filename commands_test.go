package gomander

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCustomCommandExecutesWithArgumentsFlagsAndNestedCommands(t *testing.T) {
	var gotArgument string
	var verbose bool

	child := &cobra.Command{
		Use:  "inspect target",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gotArgument = args[0]
			return nil
		},
	}
	child.Flags().BoolVar(&verbose, "verbose", false, "show verbose output")

	tools := &cobra.Command{Use: "tools"}
	tools.AddCommand(child)

	config := configWithOptions(WithCommands(tools))
	rootCmd, err := buildRootCommand(config)
	if err != nil {
		t.Fatalf("buildRootCommand() error = %v", err)
	}

	rootCmd.SetArgs([]string{"tools", "inspect", "server", "--verbose"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if gotArgument != "server" {
		t.Fatalf("custom command argument = %q, want %q", gotArgument, "server")
	}
	if !verbose {
		t.Fatal("custom command flag verbose = false, want true")
	}
}

func TestWithCommandsAccumulatesAlongsideBuiltInCommands(t *testing.T) {
	config := configWithOptions(
		WithCommands(&cobra.Command{Use: "migrate"}),
		WithCommands(&cobra.Command{Use: "version"}),
	)

	rootCmd, err := buildRootCommand(config)
	if err != nil {
		t.Fatalf("buildRootCommand() error = %v", err)
	}

	for _, name := range []string{
		"start",
		"stop",
		"restart",
		"reload",
		"status",
		"migrate",
		"version",
	} {
		command, _, err := rootCmd.Find([]string{name})
		if err != nil {
			t.Errorf("Find(%q) error = %v", name, err)
			continue
		}
		if command.Name() != name {
			t.Errorf("Find(%q) returned %q", name, command.Name())
		}
	}
}

func TestBuildRootCommandRejectsInvalidRegistrations(t *testing.T) {
	tests := []struct {
		name     string
		commands []*cobra.Command
		want     string
	}{
		{
			name:     "nil command",
			commands: []*cobra.Command{nil},
			want:     "is nil",
		},
		{
			name:     "empty command name",
			commands: []*cobra.Command{{}},
			want:     "empty name",
		},
		{
			name:     "built-in name",
			commands: []*cobra.Command{{Use: "start"}},
			want:     `token "start"`,
		},
		{
			name:     "default help name",
			commands: []*cobra.Command{{Use: "help"}},
			want:     `token "help"`,
		},
		{
			name:     "alias conflicts with built-in",
			commands: []*cobra.Command{{Use: "health", Aliases: []string{"status"}}},
			want:     `token "status"`,
		},
		{
			name: "duplicate custom names",
			commands: []*cobra.Command{
				{Use: "migrate"},
				{Use: "migrate"},
			},
			want: `token "migrate"`,
		},
		{
			name: "custom alias conflicts with name",
			commands: []*cobra.Command{
				{Use: "deploy", Aliases: []string{"ship"}},
				{Use: "ship"},
			},
			want: `token "ship"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := configWithOptions(WithCommands(test.commands...))
			rootCmd, err := buildRootCommand(config)
			if err == nil {
				t.Fatalf("buildRootCommand() error = nil, want containing %q", test.want)
			}
			if rootCmd != nil {
				t.Fatal("buildRootCommand() command is non-nil for invalid registration")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("buildRootCommand() error = %q, want containing %q", err, test.want)
			}
		})
	}
}

func TestDaemonFlagStateIsIsolatedPerConfig(t *testing.T) {
	firstConfig := defaultConfig()
	secondConfig := defaultConfig()

	firstStart := createStartCommand(firstConfig)
	createStartCommand(secondConfig)

	if err := firstStart.Flags().Set("daemon", "true"); err != nil {
		t.Fatalf("set daemon flag: %v", err)
	}

	if !firstConfig.daemonMode {
		t.Fatal("first config daemonMode = false, want true")
	}
	if secondConfig.daemonMode {
		t.Fatal("second config daemonMode = true, want false")
	}
}

func configWithOptions(options ...Option) *Config {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}
	return config
}
