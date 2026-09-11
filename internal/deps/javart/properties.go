package javart

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// Properties is a parsed .properties file. Iteration order is not defined,
// exactly as for java.util.Properties.
type Properties map[string]string

// LoadProperties parses the java.util.Properties format:
//
//   - blank lines and lines whose first non-blank character is '#' or '!' are
//     comments;
//   - a key ends at the first unescaped '=', ':' or run of whitespace;
//   - whitespace around the separator is discarded, so
//     "greeting.morning = Good morning! " has key "greeting.morning" and
//     value "Good morning! " — the trailing space survives, the leading one
//     does not;
//   - a line ending in an odd number of backslashes continues onto the next;
//   - \uXXXX, \t, \n, \r, \f and \\ are decoded in both key and value.
func LoadProperties(r io.Reader) (Properties, error) {
	props := Properties{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var pending string
	for sc.Scan() {
		line := sc.Text()
		if pending == "" {
			line = strings.TrimLeft(line, " \t\f")
			if line == "" || line[0] == '#' || line[0] == '!' {
				continue
			}
		} else {
			line = strings.TrimLeft(line, " \t\f")
		}
		if hasContinuation(line) {
			pending += strings.TrimRight(line, "\\")
			continue
		}
		logical := pending + line
		pending = ""
		key, value := splitProperty(logical)
		props[unescape(key)] = unescape(value)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if pending != "" {
		key, value := splitProperty(pending)
		props[unescape(key)] = unescape(value)
	}
	return props, nil
}

// hasContinuation reports whether the line ends in an odd number of
// backslashes, which is how .properties marks a continued line.
func hasContinuation(line string) bool {
	n := 0
	for i := len(line) - 1; i >= 0 && line[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

// splitProperty cuts a logical line into its key and value at the first
// unescaped separator.
func splitProperty(line string) (key, value string) {
	i := 0
	for ; i < len(line); i++ {
		c := line[i]
		if c == '\\' {
			i++
			continue
		}
		if c == '=' || c == ':' || c == ' ' || c == '\t' || c == '\f' {
			break
		}
	}
	if i >= len(line) {
		return line, ""
	}
	key = line[:i]
	rest := strings.TrimLeft(line[i:], " \t\f")
	if rest != "" && (rest[0] == '=' || rest[0] == ':') {
		rest = strings.TrimLeft(rest[1:], " \t\f")
	}
	return key, rest
}

// unescape decodes the escape sequences .properties allows.
func unescape(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 'f':
			b.WriteByte('\f')
		case 'u':
			if i+4 < len(s) {
				if n, err := strconv.ParseUint(s[i+1:i+5], 16, 32); err == nil {
					b.WriteRune(rune(n))
					i += 4
					continue
				}
			}
			b.WriteByte('u')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
