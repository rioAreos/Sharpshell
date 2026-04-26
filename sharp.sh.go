package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Primary  string
	Accent   string
	Location *time.Location
}

var currentStyle = Config{
	Primary:  "\033[36m", 
	Accent:   "\033[35m", 
	Location: time.Local, 
}

const (
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Bold   = "\033[1m"
	Reset  = "\033[0m"
)

func getBanner() string {
	return `
    __
   //\\        ` + currentStyle.Primary + `SHARP-shell` + Reset + `
  //  \\        @Rioareos (creator)
 //    \\        ` + currentStyle.Accent + `<-- :3 -->` + Reset + `
//      \\        ˃ 𖥦 ˂
\\______//
< ` + currentStyle.Primary + `SHARP-shell` + Reset + ` >
`
}

func notify(msg string, nType string) {
	var icon string
	switch nType {
	case "err":
		icon = Red + "[✖] "
	case "warn":
		icon = Yellow + "[!] "
	case "success":
		icon = Green + "[✔] "
	default:
		icon = currentStyle.Primary + "[⚡] "
	}
	fmt.Printf("%s%s%s\n", icon, msg, Reset)
}

func handleManage() {
	fmt.Println(Bold + "\n--- SHARP MANAGEMENT ---" + Reset)
	fmt.Printf("Local Time: %s%s%s\n", Yellow, time.Now().In(currentStyle.Location).Format("15:04:05"), Reset)
	fmt.Println("1. [MAINTENANCE] | Sleep for X minutes/seconds")
	fmt.Println("2. [SHARP STYLE] | Theme settings")
	fmt.Println("3. [SHARP RESET] | Factory reset")
	fmt.Print("\nSelect option (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("Enter duration (e.g., 5m, 30s, 1m30s): ")
		durStr, _ := reader.ReadString('\n')
		durStr = strings.TrimSpace(strings.ToLower(durStr))

		// Parse the duration (Go handles m, s, h automatically)
		duration, err := time.ParseDuration(durStr)
		if err != nil {
			notify("Invalid format! Use 'm' for minutes and 's' for seconds (e.g. 10m).", "err")
			return
		}

		notify(fmt.Sprintf("Maintenance active. Sleeping for %v...", duration), "warn")
		
		// Start the sleep
		time.Sleep(duration)
		
		notify("Maintenance complete! Sharp-shell is awake. ˃ 𖥦 ˂", "success")

	case "2":
		fmt.Println("\nAvailable Themes:")
		fmt.Println("A. Sakura Pink")
		fmt.Println("B. Neon Green")
		fmt.Println("C. Sky Blue")
		fmt.Print("Pick one: ")
		pick, _ := reader.ReadString('\n')
		pick = strings.TrimSpace(strings.ToUpper(pick))
		if pick == "A" { currentStyle.Primary = "\033[38;5;205m"; currentStyle.Accent = "\033[38;5;213m" }
		if pick == "B" { currentStyle.Primary = "\033[38;5;82m"; currentStyle.Accent = "\033[38;5;154m" }
		if pick == "C" { currentStyle.Primary = "\033[36m"; currentStyle.Accent = "\033[35m" }
		notify("Style updated!", "success")

	case "3":
		currentStyle.Primary = "\033[36m"
		currentStyle.Accent = "\033[35m"
		fmt.Print("\033[H\033[2J")
		fmt.Print(getBanner())
		notify("Reset complete.", "success")
	}
}

// ... (Other handlers like handlePkg and handleSecure remain the same) ...

func handlePkg(action string, args []string) {
	fullArgs := append([]string{action}, args...)
	cmd := exec.Command("pkg", fullArgs...)
	var buf bytes.Buffer
	multiWriter := io.MultiWriter(os.Stdout, &buf)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = multiWriter, multiWriter, os.Stdin
	err := cmd.Run()
	output := strings.ToLower(buf.String())
	if err != nil {
		notify("Dangerous problem occured!", "err")
	} else if strings.Contains(output, "warning") || strings.Contains(output, "already installed") {
		notify("Pkg "+action+" done with warnings.", "warn")
	} else {
		notify("Pkg "+action+" success! ˃ 𖥦 ˂", "success")
	}
}

func handleSecure(args []string) {
	if len(args) < 4 { notify("Usage: --secure clang <file> -o <name>", "warn"); return }
	compiler, srcFile, outputName := args[0], args[1], args[3]
	homeDir := os.Getenv("HOME")
	srcPath, _ := filepath.Abs(srcFile)
	destSrc := filepath.Join(homeDir, filepath.Base(srcFile))
	destBin := filepath.Join(homeDir, outputName)
	input, _ := os.ReadFile(srcPath)
	os.WriteFile(destSrc, input, 0644)
	notify("Compiling in $HOME zone...", "success")
	compileCmd := exec.Command(compiler, destSrc, "-o", destBin)
	compileCmd.Stdout, compileCmd.Stderr = os.Stdout, os.Stderr
	if err := compileCmd.Run(); err == nil {
		notify("Running binary...", "success")
		runCmd := exec.Command(destBin)
		runCmd.Stdout, runCmd.Stderr, runCmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		runCmd.Run()
	} else { notify("Compilation failed!", "err") }
}

func main() {
	fmt.Print("\033[H\033[2J") 
	fmt.Print(getBanner())
	
	scanner := bufio.NewScanner(os.Stdin)

	for {
		cwd, _ := os.Getwd()
		fmt.Printf("%sSHARP %s%s%s $ ", Green, Yellow, cwd, Reset)

		if !scanner.Scan() { break }
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 { continue }

		command := parts[0]
		args := parts[1:]

		if command == "sharp" {
			if len(args) > 0 {
				if args[0] == "--ver" {
					fmt.Println(currentStyle.Primary + Bold + "SHARP-shell v1.4" + Reset)
					fmt.Println("Creator:   @Rioareos")
					fmt.Printf("Local:     %s\n", time.Now().Format("15:04:05"))
					fmt.Println("Status:    Duration-Mode Active ˃ 𖥦 ˂")
				} else if args[0] == "--manage" {
					handleManage()
				}
			}
			continue
		}

		switch command {
		case "exit":
			notify("Goodbye! ˃ 𖥦 ˂", "success")
			return
		case "--secure":
			handleSecure(args)
		case "cd":
			target := os.Getenv("HOME")
			if len(args) >= 1 { target = args[0] }
			if err := os.Chdir(target); err != nil { notify(err.Error(), "err") } else { notify("Moved.", "success") }
		case "pkg":
			if len(args) > 0 && (args[0] == "install" || args[0] == "remove") {
				handlePkg(args[0], args[1:])
			} else {
				cmd := exec.Command("pkg", args...)
				cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
				cmd.Run()
			}
		case "clear":
			fmt.Print("\033[H\033[2J")
			fmt.Print(getBanner())
		default:
			cmd := exec.Command(command, args...)
			cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
			if err := cmd.Run(); err != nil { notify("Command failed.", "err") }
		}
	}
}
