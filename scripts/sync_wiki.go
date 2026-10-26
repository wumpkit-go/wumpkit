package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type Page struct {
	Title       string
	Section     string
	Weight      int
	RelPath     string
	WikiName    string
	Content     string
}

var frontmatterRegex = regexp.MustCompile(`(?s)^---\r?\n(.*?)\r?\n---\r?\n?`)
var titleRegex = regexp.MustCompile(`(?m)^title:\s*["']?(.*?)["']?$`)
var weightRegex = regexp.MustCompile(`(?m)^weight:\s*(\d+)$`)

func main() {
	contentDir := flag.String("content", "docs/content", "path to hugo content directory")
	targetDir := flag.String("target", "", "path to cloned wiki git repository")
	dryRun := flag.Bool("dry-run", false, "dry run without writing files")
	flag.Parse()

	if *targetDir == "" && !*dryRun {
		fmt.Println("error: target directory required (or use --dry-run)")
		os.Exit(1)
	}

	pages, sections, err := scanContent(*contentDir)
	if err != nil {
		fmt.Printf("error scanning content: %v\n", err)
		os.Exit(1)
	}

	sidebarContent := generateSidebar(pages, sections)

	if *dryRun {
		fmt.Printf("scanned %d pages across %d sections\n", len(pages), len(sections))
		fmt.Println("\n--- generated _Sidebar.md ---")
		fmt.Println(sidebarContent)
		return
	}

	if err := os.MkdirAll(*targetDir, 0755); err != nil {
		fmt.Printf("failed creating target dir: %v\n", err)
		os.Exit(1)
	}

	for _, p := range pages {
		destPath := filepath.Join(*targetDir, p.WikiName+".md")
		if err := os.WriteFile(destPath, []byte(p.Content), 0644); err != nil {
			fmt.Printf("failed writing %s: %v\n", destPath, err)
			os.Exit(1)
		}
		fmt.Printf("synced: %s -> %s.md\n", p.RelPath, p.WikiName)
	}

	sidebarPath := filepath.Join(*targetDir, "_Sidebar.md")
	if err := os.WriteFile(sidebarPath, []byte(sidebarContent), 0644); err != nil {
		fmt.Printf("failed writing _Sidebar.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("synced: _Sidebar.md")
}

func scanContent(root string) ([]Page, []string, error) {
	var pages []Page
	sectionsSet := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return err
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		cleanContent, title, weight := parseMarkdown(string(raw), info.Name())

		section := ""
		dir := filepath.Dir(rel)
		if dir != "." {
			section = cleanSectionName(dir)
			sectionsSet[section] = true
		}

		wikiName := toWikiName(rel, info.Name(), section)

		pages = append(pages, Page{
			Title:    title,
			Section:  section,
			Weight:   weight,
			RelPath:  rel,
			WikiName: wikiName,
			Content:  cleanContent,
		})

		return nil
	})

	var sections []string
	for s := range sectionsSet {
		sections = append(sections, s)
	}
	sort.Strings(sections)

	return pages, sections, err
}

func parseMarkdown(raw, fileName string) (string, string, int) {
	title := strings.TrimSuffix(fileName, ".md")
	weight := 99

	cleaned := raw
	if frontmatterRegex.MatchString(raw) {
		match := frontmatterRegex.FindStringSubmatch(raw)
		fm := match[1]
		cleaned = raw[len(match[0]):]

		if tm := titleRegex.FindStringSubmatch(fm); len(tm) > 1 {
			title = tm[1]
		}
		if wm := weightRegex.FindStringSubmatch(fm); len(wm) > 1 {
			fmt.Sscanf(wm[1], "%d", &weight)
		}
	}

	return cleaned, title, weight
}

func cleanSectionName(dir string) string {
	parts := strings.Split(dir, string(filepath.Separator))
	name := parts[0]
	name = strings.ReplaceAll(name, "-", " ")
	return toTitle(name)
}

func toWikiName(rel, fileName, section string) string {
	if fileName == "_index.md" {
		if section == "" {
			return "Home"
		}
		return strings.ReplaceAll(section, " ", "-")
	}
	name := strings.TrimSuffix(fileName, ".md")
	name = strings.ReplaceAll(name, "-", " ")
	name = toTitle(name)
	return strings.ReplaceAll(name, " ", "-")
}

func toTitle(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func generateSidebar(pages []Page, sections []string) string {
	var sb strings.Builder
	sb.WriteString("### [Overview](Home)\n\n")

	for _, sec := range sections {
		sb.WriteString(fmt.Sprintf("### %s\n", sec))
		var secPages []Page
		for _, p := range pages {
			if p.Section == sec && !strings.HasSuffix(p.RelPath, "_index.md") {
				secPages = append(secPages, p)
			}
		}
		sort.Slice(secPages, func(i, j int) bool {
			if secPages[i].Weight != secPages[j].Weight {
				return secPages[i].Weight < secPages[j].Weight
			}
			return secPages[i].Title < secPages[j].Title
		})

		for _, p := range secPages {
			sb.WriteString(fmt.Sprintf("- [%s](%s)\n", p.Title, p.WikiName))
		}
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String()) + "\n"
}
