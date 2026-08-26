package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: now <comando> [opções]")
		fmt.Println("Comandos disponíveis: init, serve, run, g")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit()
	case "serve", "dev":
		fmt.Println("🚀 Iniciando servidor de desenvolvimento...")
	case "run", "build":
		fmt.Println("📦 Compilando e enviando para produção...")
	default:
		fmt.Printf("Comando '%s' não reconhecido.\n", command)
	}
}

func runInit() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- Configuração do nowTS ---")

	// 1. Escolha do Alvo
	fmt.Println("Selecione o alvo de deploy:")
	fmt.Println("1) VPS Linux (SSH/Caddy/Nginx)")
	fmt.Println("2) cPanel / Shared Hosting (SFTP)")
	fmt.Print("Opção (1 ou 2): ")
	
	targetInput, _ := reader.ReadString('\n')
	targetInput = strings.TrimSpace(targetInput)

	target := "vps"
	if targetInput == "2" {
		target = "cpanel"
	}

	// 2. Input do Host
	fmt.Print("Host / Domínio do Servidor (ex: app.meusite.com): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)

	fmt.Printf("\n✅ Projeto configurado para %s no host %s!\n", target, host)
}
