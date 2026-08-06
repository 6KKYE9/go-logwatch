package main

import "testing"

func TestLevelOf(t *testing.T) {
	cases := map[string]string{
		"[ERROR] boom":        "ERROR",
		"fatal: down":        "ERROR",
		"warning: low":       "WARN",
		"INFO starting":      "INFO",
		"debug trace":        "DEBUG",
		"just some text":     "OTHER",
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
	// 保留命中 a 的
	got, err := filterLines(lines, "a", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "abc" || got[1] != "aXX" {
		t.Fatalf("keep 模式不符: %#v", got)
	}
	// 反向只留没命中的
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
