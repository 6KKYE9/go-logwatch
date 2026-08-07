package main

import (
	"bufio"
	"encoding/json"
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

// 统计每行里出现频次最高的词（按非空白 token），top N
func topWords(lines []string, n int) []struct {
	Word  string
	Count int
} {
	counts := map[string]int{}
	for _, l := range lines {
		for _, w := range strings.Fields(l) {
			w = strings.Trim(strings.ToLower(w), ".,:;\"'()[]{}")
			if w == "" {
				continue
			}
			counts[w]++
		}
	}
	type kv struct {
		Word  string
		Count int
	}
	pairs := make([]kv, 0, len(counts))
	for w, c := range counts {
		pairs = append(pairs, kv{w, c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count == pairs[j].Count {
			return pairs[i].Word < pairs[j].Word
		}
		return pairs[i].Count > pairs[j].Count
	})
	if n > 0 && n < len(pairs) {
		pairs = pairs[:n]
	}
	res := make([]struct {
		Word  string
		Count int
	}, len(pairs))
	for i, p := range pairs {
		res[i] = p
	}
	return res
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
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
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
	topN := flag.Int("top", 0, "统计出现最多的 N 个词（0 表示不统计）")
	jsonOut := flag.Bool("json", false, "用 JSON 输出统计结果")
	level := flag.String("level", "", "只看指定级别的行，比如 -level ERROR（大小写不敏感）")
	context := flag.Int("context", 0, "命中行（pattern 或 level）前后各多打 N 行上下文")
	flag.Parse()

	lines, err := readLines(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读日志失败: %v\n", err)
		os.Exit(1)
	}

	// level 过滤：先按级别筛出要保留的行下标
	var keepIdx map[int]bool
	if *level != "" {
		want := strings.ToUpper(*level)
		keepIdx = map[int]bool{}
		for i, l := range lines {
			if strings.Contains(strings.ToUpper(levelOf(l)), want) || strings.Contains(strings.ToUpper(l), want) {
				keepIdx[i] = true
			}
		}
	}

	// 关键词过滤优先于统计
	if *pattern != "" {
		filtered, err := filterLines(lines, *pattern, *keep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "正则不对: %v\n", err)
			os.Exit(1)
		}
		if *context > 0 {
			filtered = withContext(lines, filtered, *context)
		}
		for _, l := range filtered {
			fmt.Println(l)
		}
		return
	}

	if *level != "" {
		var out []string
		for i, l := range lines {
			if keepIdx[i] {
				out = append(out, l)
			}
		}
		if *context > 0 {
			out = withContext(lines, out, *context)
		}
		for _, l := range out {
			fmt.Println(l)
		}
		return
	}

	if *statsMode {
		stats := levelStats(lines)
		if *jsonOut {
			b, _ := json.MarshalIndent(stats, "", "  ")
			fmt.Println(string(b))
		} else {
			printStats(stats)
		}
		return
	}

	if *topN > 0 {
		for _, p := range topWords(lines, *topN) {
			fmt.Printf("%-20s %d\n", p.Word, p.Count)
		}
		return
	}

	for _, l := range lines {
		fmt.Println(l)
	}
}

// withContext 给命中的行附加前后各 n 行上下文，按原顺序去重输出
func withContext(all, hits []string, n int) []string {
	// 按内容匹配位置：同一行内容可能出现多次，要全部标出来
	hitIdx := map[int]bool{}
	for i, l := range all {
		for _, h := range hits {
			if l == h {
				hitIdx[i] = true
				break
			}
		}
	}
	seen := map[int]bool{}
	var out []string
	for i := range all {
		if !hitIdx[i] {
			continue
		}
		lo := i - n
		if lo < 0 {
			lo = 0
		}
		hi := i + n
		if hi >= len(all) {
			hi = len(all) - 1
		}
		for j := lo; j <= hi; j++ {
			if !seen[j] {
				seen[j] = true
				out = append(out, all[j])
			}
		}
	}
	return out
}
