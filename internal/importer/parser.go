package importer

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var russianMonths = map[string]int{
	"января": 1, "февраля": 2, "марта": 3, "апреля": 4,
	"мая": 5, "июня": 6, "июля": 7, "августа": 8,
	"сентября": 9, "октября": 10, "ноября": 11, "декабря": 12,
}

type ParsedSet struct {
	Weight *float64 // nil = bodyweight
	Reps   int
}

type ParsedExercise struct {
	Name string
	Sets []ParsedSet
}

type ParsedWorkout struct {
	Title     string
	Date      time.Time
	Exercises []ParsedExercise
}

var (
	reSepBlock  = regexp.MustCompile(`(?m)\n[ \t]*---+[ \t]*\n`)
	reBlankLine = regexp.MustCompile(`\n{2,}`)
	reHeaderNew = regexp.MustCompile(`(?i)^[Тт]ренировка\s+(\d{1,2})\.(\d{2})\.(\d{4})\.\s*(.+?)\.?\s*$`)
	reHeaderOld = regexp.MustCompile(`(?i)^[Тт]ренировка\s+\d+\.\s*(.+?)\.?\s*$`)
	reDateOld   = regexp.MustCompile(`(\d{1,2})\s+([а-яёА-ЯЁ]+)\s+(\d{4})`)
	reExLine    = regexp.MustCompile(`^\d+\.\s*`)
	reSetStart  = regexp.MustCompile(`(\d+[,.]?\d*\s*×|\d+\s+к\s*×|\+\d|б/в|по\s+(?:\d|рублю)|пустая|дроп\s*сет)`)
	reMul       = regexp.MustCompile(`[×xхХ✕]`) // normalise to ×
)

func Parse(text string) []ParsedWorkout {
	// normalise multiplication signs
	text = reMul.ReplaceAllString(text, "×")
	// fix "по по N"
	text = regexp.MustCompile(`по\s+по\s*(\d)`).ReplaceAllString(text, "по $1")

	var blocks []string
	if reSepBlock.MatchString(text) {
		blocks = reSepBlock.Split(text, -1)
	} else {
		blocks = reBlankLine.Split(text, -1)
	}

	var result []ParsedWorkout
	for _, b := range blocks {
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		if w := parseBlock(b); w != nil {
			result = append(result, *w)
		}
	}
	return result
}

func parseBlock(block string) *ParsedWorkout {
	lines := nonEmpty(strings.Split(block, "\n"))
	if len(lines) < 2 {
		return nil
	}

	var title string
	var date time.Time
	exStart := 1

	if m := reHeaderNew.FindStringSubmatch(lines[0]); m != nil {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])
		date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		title = strings.TrimSpace(m[4])
	} else if m := reHeaderOld.FindStringSubmatch(lines[0]); m != nil {
		title = strings.TrimSpace(m[1])
		for i := 1; i < len(lines); i++ {
			if dm := reDateOld.FindStringSubmatch(lines[i]); dm != nil {
				day, _ := strconv.Atoi(dm[1])
				monthNum := russianMonths[strings.ToLower(dm[2])]
				year, _ := strconv.Atoi(dm[3])
				if monthNum > 0 {
					date = time.Date(year, time.Month(monthNum), day, 0, 0, 0, 0, time.UTC)
					exStart = i + 1
					break
				}
			}
		}
		if date.IsZero() {
			return nil
		}
	} else {
		return nil
	}

	var exercises []ParsedExercise
	for _, line := range lines[exStart:] {
		if !reExLine.MatchString(line) {
			continue
		}
		line = reExLine.ReplaceAllString(line, "")
		name, sets := parseExerciseLine(line)
		if name != "" {
			exercises = append(exercises, ParsedExercise{Name: name, Sets: sets})
		}
	}

	return &ParsedWorkout{Title: title, Date: date, Exercises: exercises}
}

func parseExerciseLine(line string) (string, []ParsedSet) {
	loc := reSetStart.FindStringIndex(line)
	if loc == nil {
		return strings.TrimRight(strings.TrimSpace(line), "."), nil
	}
	name := strings.TrimRight(strings.TrimSpace(line[:loc[0]]), ".")
	setsText := line[loc[0]:]
	return name, parseSetsText(setsText)
}

func parseSetsText(text string) []ParsedSet {
	var sets []ParsedSet
	for _, group := range splitGroups(text) {
		sets = append(sets, parseGroup(group)...)
	}
	return sets
}

// splitGroups splits by comma but not decimal commas like "7,5".
func splitGroups(text string) []string {
	var groups []string
	cur := ""
	for i, c := range text {
		if c == ',' {
			next := ' '
			if i+1 < len(text) {
				next = rune(text[i+1])
			}
			if next >= '0' && next <= '9' {
				cur += string(c)
			} else {
				if g := strings.TrimSpace(cur); g != "" {
					groups = append(groups, g)
				}
				cur = ""
			}
		} else {
			cur += string(c)
		}
	}
	if g := strings.TrimSpace(cur); g != "" {
		groups = append(groups, g)
	}
	return groups
}

func parseGroup(g string) []ParsedSet {
	g = strings.TrimSpace(g)
	// strip trailing technique notes
	g = regexp.MustCompile(`(?i)\s*(в\s+\S+\s*отказ|отказ|оба\s+в\s+\S+|обе\s+в\s+\S+)\s*$`).ReplaceAllString(g, "")
	g = strings.TrimSpace(g)
	if g == "" {
		return nil
	}

	// "по рублю ×R[×S]"
	if regexp.MustCompile(`(?i)по\s+рублю`).MatchString(g) {
		rest := regexp.MustCompile(`(?i)по\s+рублю`).ReplaceAllString(g, "")
		rest = strings.TrimLeft(rest, "×")
		return expand(rest, nil)
	}

	// "б/в ×R[×S]"
	if m := regexp.MustCompile(`^б/в\s*×?\s*(.+)$`).FindStringSubmatch(g); m != nil {
		return expand(m[1], nil)
	}

	// "пустая [×R[×S]]"
	if regexp.MustCompile(`(?i)пустая`).MatchString(g) {
		rest := regexp.MustCompile(`(?i)(кривая\s+)?пустая`).ReplaceAllString(g, "")
		rest = strings.TrimLeft(rest, "×")
		return expand(rest, nil)
	}

	// "+N×R[×S]"
	if m := regexp.MustCompile(`^\+(\d+(?:[,\.]\d+)?)\s*×\s*(.+)$`).FindStringSubmatch(g); m != nil {
		w := parseFloat(m[1])
		return expand(m[2], w)
	}

	// "по N+руб×R[×S]"
	if m := regexp.MustCompile(`^по\s+(\d+(?:[,\.]\d+)?)\+(?:руб|рублю)\s*×\s*(.+)$`).FindStringSubmatch(g); m != nil {
		w := parseFloat(m[1])
		return expand(m[2], w)
	}

	// "по N×R[×S]"
	if m := regexp.MustCompile(`^по\s+(\d+(?:[,\.]\d+)?)\s*×\s*(.+)$`).FindStringSubmatch(g); m != nil {
		w := parseFloat(m[1])
		return expand(m[2], w)
	}

	// "N к×R[×S]"
	if m := regexp.MustCompile(`^(\d+(?:[,\.]\d+)?)\s+к\s*×\s*(.+)$`).FindStringSubmatch(g); m != nil {
		w := parseFloat(m[1])
		return expand(m[2], w)
	}

	// standard "N×R[×S]"
	if m := regexp.MustCompile(`^(\d+(?:[,\.]\d+)?)\s*×\s*(.+)$`).FindStringSubmatch(g); m != nil {
		w := parseFloat(m[1])
		return expand(m[2], w)
	}

	return nil
}

func expand(rest string, weight *float64) []ParsedSet {
	rest = strings.TrimSpace(rest)
	rest = strings.TrimLeft(rest, "×")
	rest = strings.TrimSpace(rest)
	// strip trailing prose
	rest = regexp.MustCompile(`\s+[а-яёА-ЯЁ].+$`).ReplaceAllString(rest, "")
	rest = strings.TrimRight(rest, ".")

	if rest == "" {
		return nil
	}

	// slash notation "R1/R2/R3"
	if strings.Contains(rest, "/") {
		var sets []ParsedSet
		for _, part := range strings.Split(rest, "/") {
			part = strings.TrimSpace(part)
			if m := regexp.MustCompile(`^(\d+(?:-\d+)?)\s*×\s*(\d+)$`).FindStringSubmatch(part); m != nil {
				r := firstInt(m[1])
				n, _ := strconv.Atoi(m[2])
				if n > 100 {
					n = 100
				}
				if r > 0 {
					for i := 0; i < n; i++ {
						sets = append(sets, ParsedSet{Weight: weight, Reps: r})
					}
				}
			} else if regexp.MustCompile(`^\d`).MatchString(part) {
				r := firstInt(part)
				if r > 0 {
					sets = append(sets, ParsedSet{Weight: weight, Reps: r})
				}
			}
		}
		return sets
	}

	// "R×S"
	if m := regexp.MustCompile(`^(\d+(?:-\d+)?)\s*×\s*(\d+)\s*$`).FindStringSubmatch(rest); m != nil {
		r := firstInt(m[1])
		n, _ := strconv.Atoi(m[2])
		if n > 100 {
			n = 100
		}
		if r > 0 && n > 0 {
			sets := make([]ParsedSet, n)
			for i := range sets {
				sets[i] = ParsedSet{Weight: weight, Reps: r}
			}
			return sets
		}
	}

	// just "R"
	if m := regexp.MustCompile(`^(\d+(?:-\d+)?)\s*$`).FindStringSubmatch(rest); m != nil {
		r := firstInt(m[1])
		if r > 0 {
			return []ParsedSet{{Weight: weight, Reps: r}}
		}
	}

	return nil
}

func parseFloat(s string) *float64 {
	s = strings.ReplaceAll(s, ",", ".")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func firstInt(s string) int {
	if idx := strings.Index(s, "-"); idx != -1 {
		s = s[:idx]
	}
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func nonEmpty(lines []string) []string {
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
