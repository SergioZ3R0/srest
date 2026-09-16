// srest is a TUI client for interacting remotely with the Slurm REST API.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SergioZ3R0/srest/internal/api"
	"github.com/SergioZ3R0/srest/internal/config"
	"github.com/SergioZ3R0/srest/internal/ui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "vault" {
		handleVault(os.Args[2:])
		return
	}

	cfg := config.Load()

	client := api.New(cfg.URL, cfg.JWT, cfg.Username, cfg.Insecure, cfg.AuthToken, cfg.ParseCustomHeaders())

	// If the user pinned an explicit version, use it; otherwise the UI will
	// auto-detect it.
	if cfg.APIVersion != "" {
		v, err := api.ParseVersion(cfg.APIVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "srest: %v\n", err)
			os.Exit(1)
		}
		client.SetVersion(v)
	}

	program := tea.NewProgram(
		ui.New(client),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "srest: %v\n", err)
		os.Exit(1)
	}
}

func handleVault(args []string) {
	if len(args) == 0 {
		printVaultUsage()
		return
	}

	switch args[0] {
	case "init":
		vaultInit()
	case "encrypt":
		vaultEncrypt()
	case "decrypt":
		vaultDecrypt()
	default:
		printVaultUsage()
	}
}

func printVaultUsage() {
	fmt.Println("srest vault - manage encrypted credentials")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  srest vault init       Create a new encrypted config file")
	fmt.Println("  srest vault encrypt    Encrypt an existing plain config file")
	fmt.Println("  srest vault decrypt    Decrypt and display the vault contents")
}

func vaultInit() {
	vaultPath := config.VaultPath()
	if vaultPath == "" {
		fmt.Fprintln(os.Stderr, "error: cannot determine home directory")
		os.Exit(1)
	}

	// Check if vault already exists
	if _, err := os.Stat(vaultPath); err == nil {
		fmt.Fprintf(os.Stderr, "Vault already exists: %s\n", vaultPath)
		fmt.Fprintln(os.Stderr, "Delete it first or use 'srest vault encrypt' to encrypt an existing config.")
		os.Exit(1)
	}

	fmt.Println("Create a new srest vault configuration")
	fmt.Println()

	// Get vault password
	pass1 := readPassword("Vault password: ")
	pass2 := readPassword("Confirm password: ")
	if pass1 != pass2 {
		fmt.Fprintln(os.Stderr, "error: passwords do not match")
		os.Exit(1)
	}
	if pass1 == "" {
		fmt.Fprintln(os.Stderr, "error: password cannot be empty")
		os.Exit(1)
	}

	// Get credentials
	fmt.Print("Slurm URL [http://localhost:6820]: ")
	url := readLine()
	if url == "" {
		url = "http://localhost:6820"
	}

	fmt.Print("SLURM JWT token: ")
	jwt := readLine()

	fmt.Print("Slurm username [" + currentUser() + "]: ")
	username := readLine()
	if username == "" {
		username = currentUser()
	}

	if err := config.InitVault(vaultPath, pass1, url, jwt, username); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Vault created: %s\n", vaultPath)
	fmt.Println("Run 'srest' to start (you will be prompted for the vault password).")
}

func vaultEncrypt() {
	vaultPath := config.VaultPath()
	configPath := strings.TrimSuffix(vaultPath, "config.vault") + "config"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Config file not found: %s\n", configPath)
		fmt.Fprintln(os.Stderr, "Create it first or use 'srest vault init'.")
		os.Exit(1)
	}

	// Check if vault already exists
	if _, err := os.Stat(vaultPath); err == nil {
		fmt.Fprintf(os.Stderr, "Vault already exists: %s\n", vaultPath)
		fmt.Fprintln(os.Stderr, "Delete it first to re-encrypt.")
		os.Exit(1)
	}

	fmt.Print("Vault password: ")
	pass1 := readLine()
	fmt.Print("Confirm password: ")
	pass2 := readLine()
	if pass1 != pass2 {
		fmt.Fprintln(os.Stderr, "error: passwords do not match")
		os.Exit(1)
	}

	if err := config.EncryptConfig(vaultPath, configPath, pass1); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Encrypted: %s -> %s\n", configPath, vaultPath)
	fmt.Println("You can now delete the plain config file.")
}

func vaultDecrypt() {
	vaultPath := config.VaultPath()
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Vault not found: %s\n", vaultPath)
		os.Exit(1)
	}

	fmt.Println("Vault content:")
	pass := readPassword("Vault password: ")

	plain, err := config.DecryptVault(vaultPath, pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(plain)
}

func currentUser() string {
	u, err := exec.Command("whoami").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(u))
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func readPassword(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	// On Linux, try to hide input with stty
	if _, err := exec.Command("stty", "-echo").StdinPipe(); err == nil {
		cmd := exec.Command("stty", "-echo")
		cmd.Stdin = os.Stdin
		_ = cmd.Run()
		defer func() {
			cmd = exec.Command("stty", "echo")
			cmd.Stdin = os.Stdin
			_ = cmd.Run()
		}()
	}
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
