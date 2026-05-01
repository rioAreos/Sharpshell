 package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// --- CONFIG & GLOBAL STATE ---
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
var shellPassword = "sharp"

// Upgraded Alias System
type AliasEntry struct {
	Cmd      string `json:"cmd"`
	Location string `json:"location"` // "Sharp" or "Linux"
}

var aliases = make(map[string]AliasEntry)
var aliasTrash = make(map[string]AliasEntry)

const aliasFile = ".sharp_aliases.json"
const trashFile = ".sharp_trash.json"

// --- PERSISTENCE ---

func saveAliases() {
	home, _ := os.UserHomeDir()
	d1, _ := json.Marshal(aliases); os.WriteFile(filepath.Join(home, aliasFile), d1, 0644)
	d2, _ := json.Marshal(aliasTrash); os.WriteFile(filepath.Join(home, trashFile), d2, 0644)
}

func loadAliases() {
	home, _ := os.UserHomeDir()
	if d1, err := os.ReadFile(filepath.Join(home, aliasFile)); err == nil { json.Unmarshal(d1, &aliases) }
	if d2, err := os.ReadFile(filepath.Join(home, trashFile)); err == nil { json.Unmarshal(d2, &aliasTrash) }
}

// --- LINUX BASHRC HELPERS ---

func addLinuxAlias(name, cmd string) {
	home, _ := os.UserHomeDir()
	f, _ := os.OpenFile(filepath.Join(home, ".bashrc"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(fmt.Sprintf("\nalias %s='%s'\n", name, cmd))
}

func removeLinuxAlias(name string) {
	home, _ := os.UserHomeDir()
	bashrc := filepath.Join(home, ".bashrc")
	data, err := os.ReadFile(bashrc)
	if err != nil { return }
	lines := strings.Split(string(data), "\n")
	var newLines []string
	prefix1, prefix2 := "alias "+name+"=", "alias "+name+" ="
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, prefix1) && !strings.HasPrefix(trimmed, prefix2) {
			newLines = append(newLines, line)
		}
	}
	os.WriteFile(bashrc, []byte(strings.Join(newLines, "\n")), 0644)
}

// --- UTILITIES ---

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
	var icon, prefix string
	switch nType {
	case "err": icon, prefix = Red+"[✖] ", "ERROR: "
	case "warn": icon, prefix = Yellow+"[!] ", "ATTENTION: "
	case "success": icon, prefix = Green+"[✔] ", "FINALLY: "
	default: icon, prefix = currentStyle.Primary+"[⚡] ", "INFO: "
	}
	fmt.Printf("%s%s%s%s\n", icon, Bold, prefix, Reset+msg)
}

func handleNotifyUsage() {
	fmt.Println(Bold + "\n--- SHARP NOTIFICATION SYSTEM ---" + Reset)
	notify("For successful tasks (rare for you, huh?).", "success")
	notify("For when you're about to break something.", "warn")
	notify("For when you actually broke it. Good job.", "err")
	notify("Generic system spam.", "info")
	fmt.Println()
}

// --- ALIAS MANAGER (NEW) ---

func handleAlias(input string) { // Legacy single-line add
	clean := strings.TrimPrefix(input, "alias ")
	parts := strings.SplitN(clean, "=", 2)
	if len(parts) != 2 { notify("It's 'alias name=cmd'. Is that too much to remember?", "err"); return }
	name := strings.TrimSpace(parts[0])
	val := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
	aliases[name] = AliasEntry{Cmd: val, Location: "Sharp"}
	saveAliases()
	notify(fmt.Sprintf("I'll remember '%s' in Sharp memory.", name), "success")
}

func handleGetAlias() {
	fmt.Println(Bold + "\n--- ALIAS MANAGER ---" + Reset)
	if len(aliases) == 0 {
		fmt.Println("No aliases found. Your life is unoptimized.")
	} else {
		fmt.Printf("%-15s | %-10s | %s\n", "NAME", "LOCATION", "COMMAND")
		fmt.Println(strings.Repeat("-", 45))
		for name, entry := range aliases {
			locColor := Yellow; if entry.Location == "Linux" { locColor = currentStyle.Primary }
			fmt.Printf("%s%-15s%s | %s%-10s%s | %s\n", Bold, name, Reset, locColor, entry.Location, Reset, entry.Cmd)
		}
	}
	
	fmt.Println("\n1. Add | 2. Remove | Press Enter to exit")
	fmt.Print("Choice: ")
	reader := bufio.NewReader(os.Stdin)
	opt, _ := reader.ReadString('\n')
	opt = strings.TrimSpace(opt)

	switch opt {
	case "1":
		fmt.Print("Name (e.g. ll): "); name, _ := reader.ReadString('\n'); name = strings.TrimSpace(name)
		fmt.Print("Command (e.g. ls -la): "); cmd, _ := reader.ReadString('\n'); cmd = strings.TrimSpace(cmd)
		fmt.Print("Location? (1: Sharp, 2: Linux): "); loc, _ := reader.ReadString('\n'); loc = strings.TrimSpace(loc)
		
		if loc == "2" {
			addLinuxAlias(name, cmd)
			aliases[name] = AliasEntry{Cmd: cmd, Location: "Linux"}
			notify("Added to Linux ~/.bashrc and mapped.", "success")
		} else {
			aliases[name] = AliasEntry{Cmd: cmd, Location: "Sharp"}
			notify("Added to Sharp memory.", "success")
		}
		saveAliases()

	case "2":
		fmt.Print("Enter name to terminate: "); name, _ := reader.ReadString('\n'); name = strings.TrimSpace(name)
		if entry, exists := aliases[name]; exists {
			if entry.Location == "Linux" {
				removeLinuxAlias(name)
				delete(aliases, name)
				notify(fmt.Sprintf("Annihilated '%s' from ~/.bashrc and memory.", name), "success")
			} else {
				aliasTrash[name] = entry
				delete(aliases, name)
				notify(fmt.Sprintf("Moved '%s' to the Trash. Scared of commitment?", name), "warn")
			}
			saveAliases()
		} else {
			notify("That alias doesn't exist. Are you imagining things?", "err")
		}
	}
}

func handleGetAliasTrash(args []string) {
	if len(args) == 0 || args[0] != "-list" {
		notify("Usage: getAliasTrash -list", "warn"); return
	}
	fmt.Println(Bold + "\n--- ALIAS TRASH BIN ---" + Reset)
	if len(aliasTrash) == 0 { notify("Trash is empty. Wow, you actually cleaned up.", "info"); return }
	
	for name, entry := range aliasTrash { fmt.Printf("- %s -> %s\n", name, entry.Cmd) }
	
	fmt.Print("\nEnter name to recover (or press Enter to leave them to rot): ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	
	if entry, exists := aliasTrash[name]; exists {
		aliases[name] = entry
		delete(aliasTrash, name)
		saveAliases()
		notify("Recovered '"+name+"'. Welcome back from the dead.", "success")
	}
}

// --- CORE HANDLERS ---

func handleFetch() {
	fmt.Println()
	fmt.Println(currentStyle.Primary + "  /\\_/\\" + Reset + "    OS:      Termux / Linux")
	fmt.Println(currentStyle.Primary + " ( o.o )" + Reset + "   Shell:   SHARP-shell v2.0")
	fmt.Println(currentStyle.Primary + "  > ^ < " + Reset + "   Creator: @Rioareos")
	fmt.Printf("           Engine:  Go %s | Salt Level: Maximum\n\n", runtime.Version())
}

func handleMatrix() {
	notify("Trying to look like a hacker? Cute.", "warn")
	chars := []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ@#$%^&*")
	for i := 0; i < 40; i++ {
		line := ""
		for j := 0; j < 80; j++ { line += string(chars[rand.Intn(len(chars))]) }
		fmt.Println(Green + line + Reset)
		time.Sleep(30 * time.Millisecond)
	}
}

func handleBreathe() {
	notify("You look stressed. Calm down before you break my code.", "success")
	fmt.Print(currentStyle.Primary + "Inhale... " + Reset); time.Sleep(3 * time.Second)
	fmt.Print(Yellow + "Hold... " + Reset); time.Sleep(3 * time.Second)
	fmt.Println(Green + "Exhale... Better? ˃ 𖥦 ˂" + Reset); time.Sleep(2 * time.Second)
}

func handleDiary(args []string, readMode bool) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "sharp_diary.txt")
	if readMode {
		data, err := os.ReadFile(path)
		if err != nil { notify("Your diary is as empty as my soul.", "warn"); return }
		fmt.Println(currentStyle.Accent + "\n--- Your Deepest Regrets ---\n" + string(data) + "----------------------------\n" + Reset)
		return
	}
	if len(args) == 0 { notify("Write something. I don't have all day.", "err"); return }
	entry := strings.Join(args, " ")
	ts := time.Now().Format("2006-01-02 15:04:05")
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(fmt.Sprintf("[%s] %s\n", ts, entry))
	notify("Saved. Don't worry, I won't tell anyone... maybe.", "success")
}

func handleCalc(args []string) {
	if len(args) != 3 { notify("Format: 'calc 2 + 2'. Math isn't that hard, is it?", "err"); return }
	n1, err1 := strconv.ParseFloat(args[0], 64)
	n2, err2 := strconv.ParseFloat(args[2], 64)
	if err1 != nil || err2 != nil { notify("Those aren't numbers. Try again, genius.", "err"); return }
	var res float64
	switch args[1] {
	case "+": res = n1 + n2
	case "-": res = n1 - n2
	case "*": res = n1 * n2
	case "/": 
		if n2 == 0 { notify("Dividing by zero? You're trying to kill me.", "err"); return }
		res = n1 / n2
	default: notify("Unknown operator. I only do the basics.", "err"); return
	}
	notify(fmt.Sprintf("Result: %v (Wow, you couldn't do that in your head?)", res), "success")
}

func handlePkg(args []string) {
	notify("Invoking the package manager. Hope you know what you're doing.", "warn")
	cmd := exec.Command("pkg", args...)
	var buf bytes.Buffer
	multi := io.MultiWriter(os.Stdout, &buf)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = multi, multi, os.Stdin
	cmd.Run()
}

func handleSecure(args []string) {
	if len(args) == 0 { notify("Usage: --secure <compiler> <file> -o <name> or --secure ./bin. Read carefully.", "warn"); return }
	homeDir := os.Getenv("HOME")
	var filesToClean []string
	defer func() {
		for _, f := range filesToClean { os.Remove(f) }
		if len(filesToClean) > 0 { notify("Cleaned up your mess in $HOME.", "success") }
	}()

	if len(args) == 1 {
		abs, _ := filepath.Abs(args[0])
		dest := filepath.Join(homeDir, filepath.Base(args[0]))
		data, err := os.ReadFile(abs)
		if err != nil { notify("That file doesn't exist. Did you imagine it?", "err"); return }
		os.WriteFile(dest, data, 0755)
		filesToClean = append(filesToClean, dest)
		notify("Running binary. Stay safe.", "success")
		cmd := exec.Command(dest); cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		cmd.Run(); return
	}

	if len(args) >= 4 {
		src := args[1]; outputName := args[3]
		visited := make(map[string]bool); queue := []string{src}
		re := regexp.MustCompile(`^\s*#include\s+"([^"]+)"`)
		for len(queue) > 0 {
			curr := queue[0]; queue = queue[1:]
			if visited[curr] { continue }; visited[curr] = true
			abs, _ := filepath.Abs(curr)
			data, err := os.ReadFile(abs)
			if err != nil { continue }
			destPath := filepath.Join(homeDir, filepath.Base(curr))
			os.WriteFile(destPath, data, 0644)
			filesToClean = append(filesToClean, destPath)
			scan := bufio.NewScanner(bytes.NewReader(data))
			for scan.Scan() {
				if m := re.FindStringSubmatch(scan.Text()); len(m) > 1 { queue = append(queue, m[1]) }
			}
		}
		destBin := filepath.Join(homeDir, outputName)
		filesToClean = append(filesToClean, destBin)
		notify("Compiling safely. Try not to have syntax errors.", "warn")
		comp := exec.Command(args[0], filepath.Join(homeDir, filepath.Base(src)), "-o", destBin)
		comp.Stdout, comp.Stderr = os.Stdout, os.Stderr
		if err := comp.Run(); err == nil {
			run := exec.Command(destBin); run.Stdout, run.Stderr, run.Stdin = os.Stdout, os.Stderr, os.Stdin
			run.Run()
		} else { notify("Compilation failed. Fix your code.", "err") }
	}
}

func handleManage() {
	fmt.Println(Bold + "\n--- SHARP SETTINGS ---" + Reset)
	fmt.Println("1. Sleep | 2. Style | 3. Clear ALL Aliases")
	fmt.Print("Choice: ")
	reader := bufio.NewReader(os.Stdin)
	opt, _ := reader.ReadString('\n')
	switch strings.TrimSpace(opt) {
	case "1":
		fmt.Print("Time? (e.g. 5s): "); d, _ := reader.ReadString('\n')
		dur, _ := time.ParseDuration(strings.TrimSpace(d))
		notify("Maintenance mode. Go get a coffee.", "warn")
		time.Sleep(dur); notify("I'm back. Did you miss me?", "success")
	case "2":
		fmt.Println("A. Sakura | B. Neon | C. Classic")
		p, _ := reader.ReadString('\n')
		switch strings.ToUpper(strings.TrimSpace(p)) {
		case "A": currentStyle.Primary, currentStyle.Accent = "\033[38;5;205m", "\033[38;5;213m"
		case "B": currentStyle.Primary, currentStyle.Accent = "\033[38;5;82m", "\033[38;5;154m"
		case "C": currentStyle.Primary, currentStyle.Accent = "\033[36m", "\033[35m"
		}
		notify("New skin, same old shell.", "success")
	case "3": 
		aliases = make(map[string]AliasEntry); aliasTrash = make(map[string]AliasEntry)
		saveAliases(); notify("Memory wiped. Everything is gone.", "success")
	}
}

// --- MAIN LOOP ---

func main() {
	rand.Seed(time.Now().UnixNano()); loadAliases()
	fmt.Print("\033[H\033[2J" + getBanner())
	scanner := bufio.NewScanner(os.Stdin)

	for {
		cwd, _ := os.Getwd()
		fmt.Printf("%sSHARP %s%s%s $ ", Green, Yellow, cwd, Reset)
		if !scanner.Scan() { break }
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" { continue }

		if strings.HasPrefix(raw, "alias ") { handleAlias(raw); continue }

		parts := strings.Fields(raw)
		
		// Upgraded Expansion Logic
		if exp, ok := aliases[parts[0]]; ok {
			raw = exp.Cmd + " " + strings.Join(parts[1:], " ")
			parts = strings.Fields(raw)
		}

		commandHistory = append(commandHistory, raw)
		cmd := parts[0]; args := parts[1:]

		switch cmd {
		case "help":
			fmt.Println(Bold + "Standard Tools:" + Reset + " cd, ls, pkg, alias, getAlias, getAliasTrash, --secure, calc, history, clear, exit")
			fmt.Println(Bold + "Extra Tools:" + Reset + " fetch, weather, ping, roll, flip, matrix, breathe, pomodoro, joke, quote, lock, diary, read-diary, notify-usage")
			fmt.Println(Bold + "Special:" + Reset + " sharp --manage")
		case "version":
			fmt.Println(Bold + "\n--- SHARP-shell v2.0 ---" + Reset)
			fmt.Println(currentStyle.Primary + "[Recent Updates]" + Reset)
			fmt.Println("- Added cross-environment Alias Manager (getAlias)")
			fmt.Println("- Added Alias Trash Bin (getAliasTrash)")
			fmt.Println("- Fixed weather output so you can actually see the clouds")
			fmt.Println("- Upgraded error notifications to be 100% more condescending")
			fmt.Println()
			fmt.Println(currentStyle.Accent + "[Developer Info]" + Reset)
			fmt.Println("- I love oreo -> rio areos")
			fmt.Println("- Rioareos social acc on discord: @nahIdontSpeakGivingSocialAcc")
			fmt.Println()
		case "notify-usage": handleNotifyUsage()
		case "getAlias": handleGetAlias()
		case "getAliasTrash": handleGetAliasTrash(args)
		case "sharp":
			if len(args) > 0 && args[0] == "--manage" { handleManage() }
		case "exit": notify("Finally, some peace and quiet. Goodbye.", "success"); return
		case "clear": fmt.Print("\033[H\033[2J" + getBanner())
		case "ping": notify("PONG. Yes, I'm alive. Unlike your brain cell.", "success")
		case "fetch": handleFetch()
		case "weather": 
			city := ""; if len(args) > 0 { city = args[0] }
			notify("Fetching clouds... don't blame me if it's raining.", "info")
			_, err := exec.LookPath("curl")
			if err != nil {
				notify("You don't have 'curl' installed. Run 'pkg install curl'.", "err")
			} else {
				c := exec.Command("curl", "-s", "wttr.in/"+city+"?0")
				c.Stdout, c.Stderr = os.Stdout, os.Stderr
				if err := c.Run(); err != nil { notify("Weather service ignored you.", "err") }
			}
		case "calc": handleCalc(args)
		case "diary": handleDiary(args, false)
		case "read-diary": handleDiary(nil, true)
		case "history":
			for i, h := range commandHistory { fmt.Printf("%d: %s\n", i+1, h) }
		case "roll":
			m := 6; if len(args) > 0 { m, _ = strconv.Atoi(args[0]) }
			notify(fmt.Sprintf("You rolled a %d. Luck won't save you here.", rand.Intn(m)+1), "success")
		case "flip":
			side := "TAILS"; if rand.Intn(2) == 0 { side = "HEADS" }
			notify("Landed on "+side+". 50/50, just like your code working.", "success")
		case "matrix": handleMatrix()
		case "breathe": handleBreathe()
		case "pomodoro":
			notify("Starting timer. Focus for once.", "warn"); time.Sleep(25 * time.Minute); notify("Break time.", "success")
		case "joke":
			j := []string{"Your code.", "Why do programmers wear glasses? Because they can't C#."}
			notify(j[rand.Intn(len(j))], "success")
		case "quote":
			q := []string{"'It works on my machine.' - Every dev ever.", "'Solve first, code second.'"}
			notify(q[rand.Intn(len(q))], "success")
		case "lock":
			fmt.Print("\033[H\033[2J" + Red + "SHELL LOCKED." + Reset + " Password: ")
			for scanner.Scan() { if scanner.Text() == shellPassword { break }; fmt.Print("Wrong. Try again: ") }
			notify("Unlocked. Don't forget it again.", "success")
		case "pkg": handlePkg(args)
		case "--secure": handleSecure(args)
		case "cd":
			target := os.Getenv("HOME"); if len(args) > 0 { target = args[0] }
			if err := os.Chdir(target); err != nil { notify("I can't go to '"+target+"'. Does it exist?", "err") } else {
				newDir, _ := os.Getwd(); os.Setenv("PWD", newDir)
			}
		default:
			c := exec.Command(cmd, args...); c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
			if err := c.Run(); err != nil { notify("I don't know what '"+cmd+"' is. Are you speaking gibberish?", "err") }
		}
	}
}
