package main

import (
	"strings"
	"testing"
)

func TestLevelOf(t *testing.T) {
	cases := map[string]string{
		"[ERROR] boom":   "ERROR",
		"fatal: down":    "ERROR",
		"warning: low":   "WARN",
		"INFO starting":  "INFO",
		"debug trace":    "DEBUG",
		"just some text": "OTHER",
	}
	for line, want := range cases {
		if got := levelOf(line); got != want {
			t.Fatalf("levelOf(%q)=%s 想要 %s", line, got, want)
		}
	}
}

func TestLevelStats(t *testing.T) {
	lines := []string{"ERROR a", "warn b", "INFO c", "info d", ""}
	stats := levelStats(lines)
	if stats["ERROR"] != 1 || stats["WARN"] != 1 || stats["INFO"] != 2 {
		t.Fatalf("统计不对: %#v", stats)
	}
}

func TestFilterLines(t *testing.T) {
	lines := []string{"abc", "def", "aXX", "ghi"}
	got, err := filterLines(lines, "a", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "abc" || got[1] != "aXX" {
		t.Fatalf("keep 模式不符: %#v", got)
	}
	got, _ = filterLines(lines, "a", false)
	if len(got) != 2 || got[0] != "def" {
		t.Fatalf("反向模式不符: %#v", got)
	}
}

func TestFilterLinesBadRegex(t *testing.T) {
	if _, err := filterLines([]string{"x"}, "(", true); err == nil {
		t.Fatal("非法正则应报错")
	}
}

func TestTopWords(t *testing.T) {
	lines := []string{"error foo bar", "error foo foo", "bar baz"}
	top := topWords(lines, 2)
	if len(top) != 2 {
		t.Fatalf("top 数量不符: %#v", top)
	}
	if top[0].Word != "foo" || top[0].Count != 3 {
		t.Fatalf("高频词应为 foo(3)，实际 %#v", top)
	}
}

func TestWithContext(t *testing.T) {
	all := []string{"a", "b", "ERROR x", "d", "e"}
	hits := []string{"ERROR x"}
	// 命中行前后各 1 行：b / ERROR x / d
	got := withContext(all, hits, 1)
	want := []string{"b", "ERROR x", "d"}
	if len(got) != len(want) {
		t.Fatalf("上下文行数不符: %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("上下文不符: %#v vs %#v", got, want)
		}
	}
}

// 内容重复的行，应该各自独立，不能因为内容一样就少取上下文
func TestWithContextDuplicateLines(t *testing.T) {
	all := []string{"a", "ERROR", "b", "ERROR", "c"}
	hits := []string{"ERROR"} // 匹配第 1 和第 3 行（下标 1 和 3）
	got := withContext(all, hits, 1)
	// 第一处命中：a, ERROR, b；第二处：ERROR, c（b 已输出过不重复）
	want := []string{"a", "ERROR", "b", "ERROR", "c"}
	if len(got) != len(want) {
		t.Fatalf("上下文行数不符: 想要 %d 行, 实际 %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 行不符: 想要 %q, 实际 %q", i, want[i], got[i])
		}
	}
}

func TestLevelFilter(t *testing.T) {
	// 直接验证 level 过滤的命中逻辑：只保留含 ERROR 级别的行
	lines := []string{"INFO ok", "ERROR boom", "warn low", "FATAL crash", "debug trace"}
	want := strings.ToUpper("ERROR")
	var out []string
	for _, l := range lines {
		if strings.Contains(strings.ToUpper(levelOf(l)), want) || strings.Contains(strings.ToUpper(l), want) {
			out = append(out, l)
		}
	}
	if len(out) != 2 || out[0] != "ERROR boom" || out[1] != "FATAL crash" {
		t.Fatalf("level 过滤不符: %#v", out)
	}
}
