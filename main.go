package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	styleGreen  = "\x1b[32m"
	styleRed    = "\x1b[31m"
	styleYellow = "\x1b[33m"
	styleReset  = "\x1b[0m"
)

var version = "dev"

func hideCursor() { fmt.Print("\x1b[?25l") }
func showCursor() { fmt.Print("\x1b[?25h") }

func isSleepDisabled() *bool {
	out, err := exec.Command("/usr/bin/pmset", "-g").Output()
	if err != nil {
		return nil
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.ToLower(fields[0]) == "sleepdisabled" {
			v := fields[1] == "1"
			return &v
		}
	}
	f := false
	return &f
}

const pmsetPath = "/usr/bin/pmset"

func setSleepDisabled(disabled bool) error {
	val := "0"
	if disabled {
		val = "1"
	}
	args := []string{pmsetPath, "-a", "disablesleep", val}

	if os.Geteuid() == 0 {
		return exec.Command(args[0], args[1:]...).Run()
	}

	cmd := exec.Command("/usr/bin/sudo", append([]string{"-n"}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s (re-run install.sh to set up passwordless sudo, or use `sudo coffee`)", msg)
	}
	return nil
}

func parseTimeout(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("duration is required")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	return d, nil
}

func formatCountdown(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func clearLine() { fmt.Print("\r\x1b[2K") }

func printUsage() {
	fmt.Print(`coffee — keep your Mac awake by toggling macOS sleep prevention.

Usage:
  coffee                Run in the foreground: enables sleep prevention and
                        keeps re-checking it (macOS can silently reset it,
                        e.g. when switching between AC and battery power),
                        then disables it again on exit (press Ctrl+C to stop).
  coffee timeout <dur>  Same as plain "coffee", but shows a live countdown
                        and stops/exits automatically after <dur>
                        (e.g. 30s, 5m, 3h).
  coffee status         Show the current sleep-prevention status.
  coffee --help         Show this help message.
  coffee --version      Show the installed version.

install.sh sets up passwordless sudo automatically, so plain "coffee"
normally just works. Without that, it only works if you already have a
cached sudo credential — otherwise run it as "sudo coffee" instead.
"status" never needs a password.
`)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printUsage()
			return
		case "-v", "--version", "version":
			fmt.Println("coffee " + version)
			return
		case "status":
			switch disabled := isSleepDisabled(); {
			case disabled == nil:
				fmt.Println(styleYellow + "● unknown" + styleReset)
			case *disabled:
				fmt.Println(styleGreen + "● sleep prevention is ON" + styleReset)
			default:
				fmt.Println(styleRed + "● sleep prevention is OFF" + styleReset)
			}
			return
		case "timeout":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Error: coffee timeout requires a duration, e.g. `coffee timeout 3h`")
				fmt.Fprintln(os.Stderr)
				printUsage()
				os.Exit(1)
			}
			d, err := parseTimeout(os.Args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
				printUsage()
				os.Exit(1)
			}
			runForeground(d)
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	runForeground(0)
}

func runForeground(timeout time.Duration) {
	if err := setSleepDisabled(true); err != nil {
		fmt.Fprintln(os.Stderr, "Error: could not enable sleep prevention:", err)
		os.Exit(1)
	}

	cleanup := func(reason string) {
		clearLine()
		showCursor()
		_ = setSleepDisabled(false)
		fmt.Println(styleRed + "✔ Sleep prevention is OFF" + reason + styleReset)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	healthTicker := time.NewTicker(30 * time.Second)
	defer healthTicker.Stop()

	var timeoutCh <-chan time.Time
	var countdownCh <-chan time.Time
	var deadline time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timeoutCh = timer.C
		deadline = time.Now().Add(timeout)

		countdownTicker := time.NewTicker(time.Second)
		defer countdownTicker.Stop()
		countdownCh = countdownTicker.C

		fmt.Println(styleGreen + "✔ Sleep prevention is ON — Ctrl+C to stop early" + styleReset)
		fmt.Printf("%s⏳ %s remaining%s", styleYellow, formatCountdown(timeout), styleReset)
	} else {
		fmt.Println(styleGreen + "✔ Sleep prevention is ON — Press Ctrl+C to stop" + styleReset)
	}
	hideCursor()
	defer showCursor()

	for {
		select {
		case <-sigCh:
			cleanup("")
			return
		case <-timeoutCh:
			cleanup(" (timeout reached)")
			return
		case <-countdownCh:
			remaining := time.Until(deadline)
			if remaining < 0 {
				remaining = 0
			}
			fmt.Printf("\r\x1b[2K%s⏳ %s remaining%s", styleYellow, formatCountdown(remaining), styleReset)
		case <-healthTicker.C:
			if disabled := isSleepDisabled(); disabled != nil && !*disabled {
				clearLine()
				fmt.Println(styleYellow + "⚠ sleep prevention was reverted externally — re-enabling" + styleReset)
				if err := setSleepDisabled(true); err != nil {
					fmt.Fprintln(os.Stderr, "Error: could not re-enable sleep prevention:", err)
				}
			}
		}
	}
}
