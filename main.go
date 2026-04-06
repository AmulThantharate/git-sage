package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"git-sage/internal/ai"
	"git-sage/internal/config"
	"git-sage/internal/detect"
	"git-sage/internal/git"
	"git-sage/internal/ui"
)

func main() {
	autoFlag := flag.Bool("auto", false, "Skip interactive picker and confirmation")
	dryRunFlag := flag.Bool("dry-run", false, "Print commit message without committing")
	noAiFlag := flag.Bool("no-ai", false, "Use rule-based message generation (not fully implemented, acts as fallback)")
	messageFlag := flag.String("message", "", "Natural language description to guide AI commit message generation")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: git-sage [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// 1. Verify Repo
	if !git.IsRepo() {
		fmt.Println("Error: Not a git repository")
		os.Exit(1)
	}

	// 2. Get Staged Diff
	diff, err := git.GetStagedDiff()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading diff: %v\n", err)
		os.Exit(1)
	}
	if diff == "" {
		fmt.Println("Error: No staged changes found. Please stage changes with 'git add'.")
		os.Exit(1)
	}

	files, err := git.GetStagedFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading staged files: %v\n", err)
		os.Exit(1)
	}

	// 3. Detect Type
	commitType, _ := detect.AnalyzeDiff(files, diff)

	// 4. Load Config
	repoRoot, _ := git.GetRepoRoot()
	cfg, err := config.Load(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
	}

	// 5. Apply branch defaults if configured
	if len(cfg.BranchDefaults) > 0 {
		if branch, err := git.GetCurrentBranch(); err == nil {
			for pattern, defaultType := range cfg.BranchDefaults {
				if strings.Contains(branch, pattern) {
					commitType = detect.CommitType(defaultType)
					break
				}
			}
		}
	}

	// AI Generator Closure — captures diff and userDescription
	userDescription := *messageFlag
	generator := func(t, e string) (string, error) {
		if *noAiFlag {
			return "update code", nil
		}
		return ai.GenerateSubject(diff, t, e, userDescription)
	}

	var finalType, finalEmoji, finalScope, finalMsg string

	if *autoFlag {
		finalType = string(commitType)
		if emoji, ok := cfg.Emojis[finalType]; ok {
			finalEmoji = emoji
		} else {
			finalEmoji = "❓"
		}

		fmt.Printf("Auto-detected: %s %s\n", finalEmoji, finalType)
		fmt.Println("Generating message...")

		msg, err := generator(finalType, finalEmoji)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating message: %v\n", err)
			os.Exit(1)
		}
		finalMsg = msg
	} else {
		t, e, scope, msg, err := ui.Run(cfg, commitType, generator)
		if err != nil {
			if err.Error() == "aborted by user" {
				fmt.Println("Aborted.")
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Error in UI: %v\n", err)
			os.Exit(1)
		}
		finalType = t
		finalEmoji = e
		finalScope = scope
		finalMsg = msg
	}

	// Build full commit message: "emoji type(scope): subject\n\nbody"
	typeStr := finalType
	if finalScope != "" {
		typeStr = fmt.Sprintf("%s(%s)", finalType, finalScope)
	}
	fullMsg := fmt.Sprintf("%s %s: %s", finalEmoji, typeStr, finalMsg)

	if *dryRunFlag {
		fmt.Printf("Dry Run: git commit -m \"%s\"\n", fullMsg)
		return
	}

	if err := git.Commit(fullMsg); err != nil {
		fmt.Fprintf(os.Stderr, "Error committing: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Commit successful!")
}
