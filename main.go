package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// 一条日志行抽出来的级别，没有就归到 OTHER
func levelOf(line string) string {
	up := strings.ToUpper(line)
	switch {
	case strings.Contains(up, "ERROR") || strings.Contains(up, "FATAL"):
		return "ERROR"
	case strings.Contains(up, "WARN"):
		return "WARN"
	case strings.Contains(up, "INFO"):
		return "INFO"
	case strings.Contains(up, "DEBUG"):
		return "DEBUG"
	default:
		return "OTHER"
	}
}

// 统计各级别出现次数
func levelStats(lines []string) map[string]int {
	stats := map[string]int{}
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		stats[levelOf(l)]++
	}
	return stats
}

// 按正则过滤行，keep 为 true 保留命中的，false 反过来只留没命中的
func filterLines(lines []string, pattern string, keep bool) ([]string, error) {
	if pattern == "" {
		if keep {
			return lines, nil
		}
		return []string{}, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, l := range lines {
		hit := re.MatchString(l)
		if hit == keep {
			out = append(out, l)
		}
	}
	return out, nil
}

func printStats(stats map[string]int) {
	order := []string{"ERROR", "WARN", "INFO", "DEBUG", "OTHER"}
	for _, k := range order {
		if v, ok := stats[k]; ok {
			fmt.Printf("%-6s %d\n", k, v)
		}
	}
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		found := false
		for _, o := range order {
			if o == k {
				found = true
			}
		}
		if !found {
			fmt.Printf("%-6s %d\n", k, stats[k])
		}
	}
}

func readLines(path string) ([]string, error) {
	var lines []string
	var sc *bufio.Scanner
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		sc = bufio.NewScanner(f)
	} else {
		sc = bufio.NewScanner(os.Stdin)
	}
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func main() {
	statsMode := flag.Bool("stats", false, "统计各日志级别出现次数")
	pattern := flag.String("pattern", "", "按正则过滤行")
	keep := flag.Bool("keep", true, "pattern 模式下保留命中的（false 则反向）")
	path := flag.String("f", "", "日志文件路径，不给就从标准输入读")
	flag.Parse()

	lines, err := readLines(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读日志失败: %v\n", err)
		os.Exit(1)
	}

	if *pattern != "" {
		filtered, err := filterLines(lines, *pattern, *keep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "正则不对: %v\n", err)
			os.Exit(1)
		}
		for _, l := range filtered {
			fmt.Println(l)
		}
		return
	}

	if *statsMode {
		printStats(levelStats(lines))
		return
	}

	for _, l := range lines {
		fmt.Println(l)
	}
}
