package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Global Config & State
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

var commandHistory []string
var shellPassword = "sharp" // Default password for the lock command

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

// ---------------------------------------------------------
// NEW COMMAND HANDLERS
// ---------------------------------------------------------

func handleFetch() {
	fmt.Println()
	fmt.Println(currentStyle.Primary + "  /\\_/\\" + Reset + "    OS:      Termux / Linux")
	fmt.Println(currentStyle.Primary + " ( o.o )" + Reset + "   Shell:   SHARP-shell v1.5")
	fmt.Println(currentStyle.Primary + "  > ^ < " + Reset + "   Creator: @Rioareos")
	fmt.Printf("           Engine:  Go %s\n", runtime.Version())
	fmt.Printf("           Arch:    %s\n\n", runtime.GOARCH)
}

func handleMatrix() {
	notify("Entering the matrix...", "warn")
	chars := []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ@#$%^&*")
	for i := 0; i < 40; i++ {
		line := ""
		for j := 0; j < 80; j++ {
			line += string(chars[rand.Intn(len(chars))])
		}
		fmt.Println(Green + line + Reset)
		time.Sleep(40 * time.Millisecond)
	}
	notify("Matrix sequence complete.", "success")
}

func handleBreathe() {
	notify("Take a moment to relax...", "success")
	fmt.Print(currentStyle.Primary + "Inhale... " + Reset)
	time.Sleep(3 * time.Second)
	fmt.Print(Yellow + "Hold... " + Reset)
	time.Sleep(3 * time.Second)
	fmt.Println(Green + "Exhale... ˃ 𖥦 ˂" + Reset)
	time.Sleep(4 * time.Second)
}

func handleDiary(args []string) {
	homeDir := os.Getenv("HOME")
	diaryPath := filepath.Join(homeDir, "sharp_diary.txt")
	
	entry := strings.Join(args, " ")
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fullEntry := fmt.Sprintf("[%s] %s\n", timestamp, entry)

	f, err := os.OpenFile(diaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		notify("Could not open diary file.", "err")
		return
	}
	defer f.Close()
	f.WriteString(fullEntry)
	notify("Secret diary entry saved! 🐾", "success")
}

func handleReadDiary() {
	homeDir := os.Getenv("HOME")
	diaryPath := filepath.Join(homeDir, "sharp_diary.txt")
	data, err := os.ReadFile(diaryPath)
	if err != nil {
		notify("Diary is empty or not found.", "warn")
		return
	}
	fmt.Println(currentStyle.Accent + "\n--- Your Secret Diary ---" + Reset)
	fmt.Print(string(data))
	fmt.Println(currentStyle.Accent + "-------------------------\n" + Reset)
}

func handleCalc(args []string) {
	if len(args) != 3 {
		notify("Usage: calc <num> <op> <num> (e.g., calc 5 + 10)", "err")
		return
	}
	num1, err1 := strconv.ParseFloat(args[0], 64)
	num2, err2 := strconv.ParseFloat(args[2], 64)
	if err1 != nil || err2 != nil {
		notify("Invalid numbers!", "err")
		return
	}
	
	var result float64
	switch args[1] {
	case "+": result = num1 + num2
	case "-": result = num1 - num2
	case "*": result = num1 * num2
	case "/":
		if num2 == 0 { notify("Cannot divide by zero!", "err"); return }
		result = num1 / num2
	default:
		notify("Unknown operator. Use +, -, *, /", "err")
		return
	}
	notify(fmt.Sprintf("Result: %v", result), "success")
}

func handleWeather(args []string) {
	city := ""
	if len(args) > 0 { city = args[0] }
	cmd := exec.Command("curl", "wttr.in/"+city+"?0")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.Run()
}

func handleLock(scanner *bufio.Scanner) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println(Red + Bold + "=== SHELL LOCKED ===" + Reset)
	for {
		fmt.Print("Enter password to unlock: ")
		if !scanner.Scan() { os.Exit(0) }
		if scanner.Text() == shellPassword {
			notify("Shell unlocked! Welcome back. ˃ 𖥦 ˂", "success")
			break
		} else {
			notify("Incorrect password.", "err")
		}
	}
}

// ---------------------------------------------------------
// CORE SYSTEM
// ---------------------------------------------------------

func handleManage() {
	fmt.Println(Bold + "\n--- SHARP MANAGEMENT ---" + Reset)
	fmt.Println("1. [MAINTENANCE] | Sleep for X minutes/seconds")
	fmt.Println("2. [SHARP STYLE] | Theme settings")
	fmt.Println("3. [SHARP RESET] | Factory reset")
	fmt.Print("\nSelect option (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("Enter duration (e.g., 5m, 30s): ")
		durStr, _ := reader.ReadString('\n')
		duration, err := time.ParseDuration(strings.TrimSpace(strings.ToLower(durStr)))
		if err != nil { notify("Invalid format!", "err"); return }
		notify(fmt.Sprintf("Maintenance active. Sleeping for %v...", duration), "warn")
		time.Sleep(duration)
		notify("Maintenance complete! ˃ 𖥦 ˂", "success")
	case "2":
		fmt.Println("A. Sakura Pink | B. Neon Green | C. Sky Blue")
		fmt.Print("Pick one: ")
		pick, _ := reader.ReadString('\n')
		pick = strings.TrimSpace(strings.ToUpper(pick))
		if pick == "A" { currentStyle.Primary, currentStyle.Accent = "\033[38;5;205m", "\033[38;5;213m" }
		if pick == "B" { currentStyle.Primary, currentStyle.Accent = "\033[38;5;82m", "\033[38;5;154m" }
		if pick == "C" { currentStyle.Primary, currentStyle.Accent = "\033[36m", "\033[35m" }
		notify("Style updated!", "success")
	case "3":
		currentStyle.Primary, currentStyle.Accent = "\033[36m", "\033[35m"
		fmt.Print("\033[H\033[2J")
		fmt.Print(getBanner())
	}
}

func handlePkg(action string, args []string) {
	fullArgs := append([]string{action}, args...)
	cmd := exec.Command("pkg", fullArgs...)
	var buf bytes.Buffer
	multiWriter := io.MultiWriter(os.Stdout, &buf)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = multiWriter, multiWriter, os.Stdin
	err := cmd.Run()
	output := strings.ToLower(buf.String())
	if err != nil { notify("Dangerous problem occured!", "err")
	} else if strings.Contains(output, "warning") || strings.Contains(output, "already installed") { notify("Pkg "+action+" done with warnings.", "warn")
	} else { notify("Pkg "+action+" success! ˃ 𖥦 ˂", "success") }
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

func printHelp() {
	fmt.Println(currentStyle.Primary + Bold + "\n--- CORE COMMANDS ---" + Reset)
	fmt.Println("cd, ls, clear, pkg, exit")
	fmt.Println("sharp --manage  | Settings & Maintenance")
	fmt.Println("--secure        | Compile C/C++ bypassing sdcard restrictions")
	
	fmt.Println(currentStyle.Accent + Bold + "\n--- THE 15 NEW SHARP TOOLS ---" + Reset)
	fmt.Println("1.  ping         | Check shell life")
	fmt.Println("2.  fetch        | Cute system info display")
	fmt.Println("3.  weather      | Check weather (e.g., weather tokyo)")
	fmt.Println("4.  calc         | Math tool (e.g., calc 5 * 10)")
	fmt.Println("5.  diary        | Save a secret entry (e.g., diary I coded today!)")
	fmt.Println("6.  read-diary   | Read all your secret entries")
	fmt.Println("7.  history      | View commands typed this session")
	fmt.Println("8.  roll         | Roll a dice (e.g., roll 20)")
	fmt.Println("9.  flip         | Flip a coin (Heads/Tails)")
	fmt.Println("10. matrix       | Hacker screen effect")
	fmt.Println("11. breathe      | Guided breathing for relaxation")
	fmt.Println("12. pomodoro     | Start a 25-minute focus timer")
	fmt.Println("13. joke         | Get a programmer joke")
	fmt.Println("14. quote        | Get an inspirational quote")
	fmt.Println("15. lock         | Lock the shell (Default password: 'sharp')\n")
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator
	fmt.Print("\033[H\033[2J") 
	fmt.Print(getBanner())
	
	scanner := bufio.NewScanner(os.Stdin)

	for {
		cwd, _ := os.Getwd()
		fmt.Printf("%sSHARP %s%s%s $ ", Green, Yellow, cwd, Reset)

		if !scanner.Scan() { break }
		input := scanner.Text()
		
		// Add to history
		if strings.TrimSpace(input) != "" {
			commandHistory = append(commandHistory, input)
		}

		parts := strings.Fields(input)
		if len(parts) == 0 { continue }

		command := parts[0]
		args := parts[1:]

		if command == "sharp" {
			if len(args) > 0 {
				if args[0] == "--manage" { handleManage() }
			}
			continue
		}

		switch command {
		case "exit": notify("Goodbye! ˃ 𖥦 ˂", "success"); return
		case "help": printHelp()
		case "--secure": handleSecure(args)
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
		case "clear": fmt.Print("\033[H\033[2J"); fmt.Print(getBanner())
		
		// --- THE 15 NEW COMMANDS ---
		case "ping": notify("pong! 🐾", "success")
		case "fetch": handleFetch()
		case "weather": handleWeather(args)
		case "calc": handleCalc(args)
		case "diary": handleDiary(args)
		case "read-diary": handleReadDiary()
		case "history":
			for i, cmd := range commandHistory {
				fmt.Printf(" %d: %s\n", i+1, cmd)
			}
		case "roll":
			max := 6
			if len(args) > 0 { if val, err := strconv.Atoi(args[0]); err == nil && val > 0 { max = val } }
			notify(fmt.Sprintf("You rolled a %d!", rand.Intn(max)+1), "success")
		case "flip":
			if rand.Intn(2) == 0 { notify("Coin landed on: HEADS", "success") } else { notify("Coin landed on: TAILS", "success") }
		case "matrix": handleMatrix()
		case "breathe": handleBreathe()
		case "pomodoro":
			notify("Pomodoro started. Focus for 25 minutes! ˃ 𖥦 ˂", "warn")
			time.Sleep(25 * time.Minute)
			notify("Pomodoro complete! Take a 5-minute break.", "success")
			fmt.Print("\a") // Terminal bell sound
		case "joke":
			jokes := []string{"Why do programmers prefer dark mode? Because light attracts bugs.", "I would tell you a UDP joke, but you might not get it.", "There are 10 types of people: those who understand binary, and those who don't."}
			notify(jokes[rand.Intn(len(jokes))], "success")
		case "quote":
			quotes := []string{"Talk is cheap. Show me the code. - Linus Torvalds", "First, solve the problem. Then, write the code. - John Johnson", "Make it work, make it right, make it fast. - Kent Beck"}
			notify(quotes[rand.Intn(len(quotes))], "success")
		case "lock": handleLock(scanner)
		
		default:
			cmd := exec.Command(command, args...)
			cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
			if err := cmd.Run(); err != nil { notify("Command failed or not found.", "err") }
		}
	}
}
