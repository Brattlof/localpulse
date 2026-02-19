package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Brattlof/localpulse/app"
	"github.com/Brattlof/localpulse/config"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("LocalPulse v%s\n", version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	setupSignalHandler(cfg)

	m := app.NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`LocalPulse - Localhost Performance Checker

USAGE:
    localpulse [OPTIONS]

OPTIONS:
    -h, --help      Show this help message
    -v, --version   Show version information

KEYBOARD SHORTCUTS:
    Tab/Shift+Tab   Focus panels
    r               Refresh/scan endpoints
    s               Start load testing
    x               Stop load testing
    +/=             Increase RPS
    -               Decrease RPS
    a               Add endpoint manually
    d               Delete selected endpoint
    Enter           Toggle load testing for selected endpoint
    q/Ctrl+C        Quit

EXAMPLES:
    localpulse              Start monitoring
    localpulse --version    Show version`)
}

func setupSignalHandler(cfg *config.Config) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cfg.Save()
		os.Exit(0)
	}()
}
