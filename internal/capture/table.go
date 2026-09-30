package capture

import "strings"

func formatTable(headers []string, rows [][]string) string {
	cols := len(headers)
	w := make([]int, cols)
	for i, h := range headers {
		w[i] = visLen(h)
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if n := visLen(row[i]); n > w[i] {
				w[i] = n
			}
		}
	}
	var b strings.Builder
	writeRow(&b, headers, w)
	for _, row := range rows {
		if len(row) < cols {
			n := make([]string, cols)
			copy(n, row)
			row = n
		}
		writeRow(&b, row, w)
	}
	return b.String()
}

func writeRow(b *strings.Builder, cols []string, w []int) {
	for i, c := range cols {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(c)
		pad := w[i] - visLen(c)
		if pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
	}
	b.WriteByte('\n')
}

// 终端里中文占两格，按这个对齐 status 表。
func visLen(s string) int {
	n := 0
	for _, r := range s {
		if r <= 127 {
			n++
		} else {
			n += 2
		}
	}
	return n
}
