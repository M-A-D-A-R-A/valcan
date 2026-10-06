// Command vulcan is the Vulcan v0 CLI.
//
// M1 ships one command: `vulcan waste`, which reads run traces from Vigil and
// reports where tokens were wasted.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"vulcan/internal/vigil"
	"vulcan/internal/waste"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "vulcan:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("vulcan", flag.ExitOnError)
	baseURL := fs.String("base-url", envOr("VIGIL_BASE_URL", "http://localhost:8080"), "Vigil server URL")
	projectID := fs.String("project-id", os.Getenv("VIGIL_PROJECT_ID"), "Vigil project id (proj_...)")
	window := fs.Duration("window", 7*24*time.Hour, "look-back window")
	sizeCap := fs.Int("size-cap", waste.DefaultSizeCap, "result size cap in bytes")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: vulcan <command> [flags]\n\ncommands:\n  waste   report wasted tokens across runs\n  runs    list runs\n\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if *projectID == "" {
		return fmt.Errorf("VIGIL_PROJECT_ID is required (or pass -project-id)")
	}

	cmd := fs.Arg(0)
	if cmd == "" {
		fs.Usage()
		return nil
	}

	client := vigil.New(*baseURL, *projectID)
	from := time.Now().Add(-*window)

	events, err := client.Logs(from, "run.")
	if err != nil {
		return err
	}

	runs := waste.Build(events)

	switch cmd {
	case "runs":
		printRuns(runs)
	case "waste":
		printWaste(runs, *sizeCap)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
	return nil
}

func printRuns(runs []waste.Run) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "RUN\tSTATUS\tSTEPS\tTOKENS\tSTARTED")
	for _, r := range runs {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n", r.ID, r.Status, len(r.Steps), r.TotalTokens, r.StartedAt.Format("15:04:05"))
	}
	w.Flush()
}

func printWaste(runs []waste.Run, sizeCap int) {
	findings := map[string][]waste.Finding{}
	for _, r := range runs {
		findings[r.ID] = waste.Analyze(r, sizeCap)
	}

	sum := waste.Summarize(runs, findings)

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "runs\t%d\n", sum.Runs)
	fmt.Fprintf(w, "steps\t%d\n", sum.Steps)
	fmt.Fprintf(w, "result bytes seen\t%d\n", sum.TotalCap)
	fmt.Fprintf(w, "addressable bytes\t%d\n", sum.Addressable)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "CLASS\tCOUNT\tBYTES")
	order := []waste.Class{waste.ClassRepeat, waste.ClassFailed, waste.ClassOversized}
	for _, c := range order {
		if sum.Findings[c] > 0 {
			fmt.Fprintf(w, "%s\t%d\t%d\n", c, sum.Findings[c], sum.Bytes[c])
		}
	}
	w.Flush()

	// Per-run detail, newest first.
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt.After(runs[j].StartedAt) })
	for _, r := range runs {
		fs := findings[r.ID]
		if len(fs) == 0 {
			continue
		}
		fmt.Printf("\n%s  (%s, %d steps)\n", r.ID, r.Status, len(r.Steps))
		for _, f := range fs {
			fmt.Printf("  %-9s %-8s %6d bytes  %s\n", f.Class, f.Step.Tool, f.Bytes, f.Detail)
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
