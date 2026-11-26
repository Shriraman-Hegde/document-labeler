package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/engine"
	"github.com/Shriraman-Hegde/document-labeler/internal/ingest"
	"github.com/Shriraman-Hegde/document-labeler/internal/llm"
	"github.com/Shriraman-Hegde/document-labeler/internal/relate"
	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

const version = "0.1.0"

// Run is the CLI entrypoint.
func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 2
	}
	switch args[0] {
	case "help", "-h", "--help":
		printHelp()
		return 0
	case "version", "--version":
		fmt.Println(version)
		return 0
	case "scan":
		return cmdScan(args[1:])
	case "accept":
		return cmdAccept(args[1:])
	case "conflicts":
		return cmdConflicts(args[1:])
	case "relate":
		return cmdRelate(args[1:])
	case "watch":
		return cmdWatch(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		printHelp()
		return 2
	}
}

func printHelp() {
	fmt.Print(`digitizer — manufacturing document digitizer

Usage:
  digitizer scan --dept <name> [--mode <mode>] [--root <dir>] <files...>
  digitizer watch --dept <name> [--once] [--root <dir>]
  digitizer accept --dept <name> --type <docType> [--root <dir>]
  digitizer conflicts --dept <name> [--root <dir>]
  digitizer relate [--dept <name>] [--root <dir>] [--all-depts]
  digitizer version

Modes: fully_learned | seeded_extensible | known_dynamic

Env:
  DIGITIZER_LLAMA_URL        llama.cpp OpenAI-compatible base URL (required)
  DIGITIZER_MODEL            model id sent to the server (default: qwen)
  DIGITIZER_LLM_TIMEOUT_SEC  HTTP timeout seconds (default: 120)
  DIGITIZER_ALLOW_STUB=1     allow heuristic stub without LLM (tests/dev)
  DIGITIZER_MODE             default mode override
`)
}

func rootFlag(args []string) (string, []string) {
	root := "."
	out := args[:0:0]
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" && i+1 < len(args) {
			root = args[i+1]
			i++
			continue
		}
		out = append(out, args[i])
	}
	return root, out
}

func deptFlag(args []string) (string, []string, error) {
	dept := ""
	out := args[:0:0]
	for i := 0; i < len(args); i++ {
		if args[i] == "--dept" && i+1 < len(args) {
			dept = args[i+1]
			i++
			continue
		}
		out = append(out, args[i])
	}
	if dept == "" {
		return "", nil, fmt.Errorf("--dept is required")
	}
	return dept, out, nil
}

func modeFlag(args []string) (schema.Mode, []string, error) {
	raw := os.Getenv("DIGITIZER_MODE")
	out := args[:0:0]
	for i := 0; i < len(args); i++ {
		if args[i] == "--mode" && i+1 < len(args) {
			raw = args[i+1]
			i++
			continue
		}
		out = append(out, args[i])
	}
	m, err := schema.ParseMode(raw)
	return m, out, err
}

func cmdScan(args []string) int {
	root, args := rootFlag(args)
	dept, args, err := deptFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	mode, args, err := modeFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "scan requires file paths")
		return 2
	}
	client, err := llm.ConfigFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := client.Ready(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfg := engine.Config{Root: root, Department: dept, ModeOverride: mode, LLM: client}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	var results []engine.Outcome
	for _, f := range args {
		o, err := engine.ProcessFile(cfg, f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			return 1
		}
		results = append(results, o)
	}
	docs, err := relate.LoadValueDir(root, dept)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rel := relate.Build(docs)
	_ = enc.Encode(map[string]any{
		"documents": results,
		"relations": rel.Relations,
		"checks":    rel.Checks,
	})
	return 0
}

func cmdRelate(args []string) int {
	root, args := rootFlag(args)
	all := false
	rest := args[:0:0]
	for _, a := range args {
		if a == "--all-depts" {
			all = true
			continue
		}
		rest = append(rest, a)
	}
	dept := ""
	if !all {
		var err error
		dept, rest, err = deptFlag(rest)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
	}
	docs, err := relate.LoadValueDir(root, dept)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	out := relate.Build(docs)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
	return 0
}

func cmdAccept(args []string) int {
	root, args := rootFlag(args)
	dept, args, err := deptFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	var typ string
	rest := args[:0:0]
	for i := 0; i < len(args); i++ {
		if args[i] == "--type" && i+1 < len(args) {
			typ = args[i+1]
			i++
			continue
		}
		rest = append(rest, args[i])
	}
	if typ == "" {
		fmt.Fprintln(os.Stderr, "--type is required")
		return 2
	}
	store := schema.Store{Root: root}
	cat, err := store.Load(dept)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := cat.Promote(schema.DocType(typ)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := store.Save(cat); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("promoted %s for department %s\n", typ, dept)
	return 0
}

func cmdConflicts(args []string) int {
	root, args := rootFlag(args)
	dept, args, err := deptFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	store := schema.Store{Root: root}
	cat, err := store.Load(dept)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	type row struct {
		Type      schema.DocType      `json:"type"`
		Version   int                 `json:"version"`
		Conflicts []schema.Conflict   `json:"conflicts"`
	}
	var rows []row
	for _, v := range cat.Versions {
		if len(v.Conflicts) == 0 {
			continue
		}
		rows = append(rows, row{Type: v.Type, Version: v.Version, Conflicts: v.Conflicts})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(rows)
	fmt.Fprintln(os.Stderr, "Review drafts/conflicts weekly (R14).")
	return 0
}

func cmdWatch(args []string) int {
	root, args := rootFlag(args)
	dept, args, err := deptFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	once := false
	for _, a := range args {
		if a == "--once" {
			once = true
		}
	}
	client, err := llm.ConfigFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := client.Ready(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	mode, args, err := modeFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	_ = args
	inbox := filepath.Join(root, "inbox", dept)
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfg := engine.Config{Root: root, Department: dept, ModeOverride: mode, LLM: client}
	processInbox := func() error {
		entries, err := os.ReadDir(inbox)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".pdf" {
				continue
			}
			path := filepath.Join(inbox, name)
			o, err := engine.ProcessFile(cfg, path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				continue
			}
			fmt.Printf("%s -> %s (%s)\n", path, o.ValuePath, o.Message)
			if !o.Rescan {
				_ = os.Remove(path) // moved conceptually to labelled via copy
			}
		}
		return nil
	}
	if err := processInbox(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if once {
		return 0
	}
	fmt.Fprintf(os.Stderr, "watching %s (Ctrl+C to stop)\n", inbox)
	if err := ingest.WatchDir(inbox, processInbox); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
