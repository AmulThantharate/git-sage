package detect

import (
	"path/filepath"
	"strings"
)

type CommitType string

const (
	Feat     CommitType = "feat"
	Fix      CommitType = "fix"
	Docs     CommitType = "docs"
	Style    CommitType = "style"
	Refactor CommitType = "refactor"
	Perf     CommitType = "perf"
	Test     CommitType = "test"
	Build    CommitType = "build"
	Ci       CommitType = "ci"
	Chore    CommitType = "chore"
	Unknown  CommitType = "unknown"
)

// AnalyzeDiff guesses the commit type based on file paths and diff content.
func AnalyzeDiff(files []string, diff string) (CommitType, float64) {
	// 1. Check for specific keywords in diff (strong signal)
	// Simple heuristic: if we see "func Test..." additions, it's likely a test
	// If we see "FIXME" removals, maybe a fix? (Hard to distinguish from adding FIXME)

	// 2. Check file extensions/names (medium signal)
	counts := make(map[CommitType]int)

	for _, f := range files {
		base := filepath.Base(f)
		ext := filepath.Ext(f)

		switch {
		case strings.HasSuffix(f, "_test.go") || strings.HasSuffix(f, ".test.js") || strings.Contains(f, "/test/"):
			counts[Test]++
		case base == "go.mod" || base == "go.sum" || base == "package.json" || base == "package-lock.json" || base == "Makefile" || base == "Dockerfile":
			counts[Build]++
		case strings.HasPrefix(base, ".github") || strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml"):
			// Could be CI or Config. But CI is more specific to workflows.
			if strings.Contains(f, "workflows") {
				counts[Ci]++
			} else {
				counts[Chore]++
			}
		case strings.HasSuffix(f, ".md") || strings.HasPrefix(base, "LICENSE") || strings.HasPrefix(base, "docs/"):
			counts[Docs]++
		case ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" || ext == ".java" || ext == ".c" || ext == ".cpp" || ext == ".rs":
			// Code files could be feat, fix, refactor, perf.
			// Let's look at the diff content for these.
			if typeFromDiff := analyzeDiffContent(diff); typeFromDiff != Unknown {
				counts[typeFromDiff] += 2 // Give more weight
			} else {
				// Default to feat if no other signal
				counts[Feat]++
			}
		default:
			counts[Chore]++
		}
	}

	// Find max
	var maxType CommitType
	var maxVal int
	for k, v := range counts {
		if v > maxVal {
			maxVal = v
			maxType = k
		}
	}

	if maxVal == 0 {
		return Feat, 0.0 // Default fallback
	}

	confidence := float64(maxVal) / float64(len(files)) // simplistic confidence
	if confidence > 1.0 {
		confidence = 1.0
	}

	return maxType, confidence
}

func analyzeDiffContent(diff string) CommitType {
	// Very basic string matching
	lowerDiff := strings.ToLower(diff)
	if strings.Contains(lowerDiff, "fix(") || strings.Contains(lowerDiff, "fixes") || strings.Contains(lowerDiff, "bug") {
		return Fix
	}
	if strings.Contains(lowerDiff, "perf") || strings.Contains(lowerDiff, "optimization") {
		return Perf
	}
	return Unknown
}
