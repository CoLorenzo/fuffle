package main

import (
	"fmt"
	"strings"
)

const (
	svgWidth      = 1200
	svgPadding    = 60
	chartHeight   = 300
	chartGap      = 80
	sectionGap    = 40
	rowHeight     = 24
	tableColWidth = 150
)

func generateSVG(eval *EvaluationFile, netdata *NetdataMetrics) string {
	var b strings.Builder

	durations := make([]float64, len(eval.Entries))
	for i, e := range eval.Entries {
		durations[i] = entryDuration(e)
	}

	mean := Mean(durations)
	std := StdDev(durations)
	median := Median(durations)
	mode := Mode(durations)
	startMin, endMax := totalDuration(eval.Entries)
	totalMs := float64(endMax - startMin)

	accuracy := calcAccuracyOverTime(eval.Entries)

	sections := 3
	if netdata != nil && len(netdata.Timestamps) > 0 {
		sections = 5
	}
	totalHeight := svgPadding + sections*(chartHeight+sectionGap) + 200

	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, svgWidth, totalHeight, svgWidth, totalHeight))
	b.WriteString(`<style>`)
	b.WriteString(`text { font-family: monospace; }`)
	b.WriteString(`.title { font-size: 24px; font-weight: bold; }`)
	b.WriteString(`.subtitle { font-size: 16px; fill: #666; }`)
	b.WriteString(`.section-title { font-size: 18px; font-weight: bold; fill: #333; }`)
	b.WriteString(`.label { font-size: 12px; fill: #666; }`)
	b.WriteString(`.value { font-size: 14px; font-weight: bold; }`)
	b.WriteString(`.ok { fill: #22c55e; }`)
	b.WriteString(`.fail { fill: #ef4444; }`)
	b.WriteString(`.grid { stroke: #e5e7eb; stroke-width: 1; }`)
	b.WriteString(`.axis { stroke: #9ca3af; stroke-width: 1; }`)
	b.WriteString(`.line-cpu { stroke: #3b82f6; stroke-width: 2; fill: none; }`)
	b.WriteString(`.line-ram { stroke: #22c55e; stroke-width: 2; fill: none; }`)
	b.WriteString(`.line-gpu { stroke: #a855f7; stroke-width: 2; fill: none; }`)
	b.WriteString(`.line-disk { stroke: #f97316; stroke-width: 2; fill: none; }`)
	b.WriteString(`.line-time { stroke: #ef4444; stroke-width: 2; fill: none; stroke-dasharray: 5,5; }`)
	b.WriteString(`.line-acc { stroke: #0ea5e9; stroke-width: 2; fill: none; }`)
	b.WriteString(`</style>`)

	y := svgPadding

	// Title
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="title">%s</text>`, svgWidth/2, y, esc(eval.Title)))
	y += 30
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="subtitle">Report generato</text>`, svgWidth/2, y))
	y += sectionGap

	// Section 1: Results table
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="section-title">Risultati</text>`, svgPadding, y))
	y += 25

	headers := []string{"File", "Stato", "Tempo (ms)", "Tag"}
	drawTableHeader(&b, svgPadding, y, headers)
	y += rowHeight

	for _, e := range eval.Entries {
		status := "OK"
		statusClass := "ok"
		if len(e.Tags) == 0 {
			status = "FAIL"
			statusClass = "fail"
		}
		tag := ""
		if len(e.Tags) > 0 {
			tag = e.Tags[0]
		}
		drawTableRow(&b, svgPadding, y, []string{e.File, status, fmt.Sprintf("%.0f", entryDuration(e)), tag}, statusClass)
		y += rowHeight
	}
	y += sectionGap

	// Section 2: Overview stats
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="section-title">Statistiche Tempo</text>`, svgPadding, y))
	y += 25
	stats := []string{
		fmt.Sprintf("Totale file: %d", len(eval.Entries)),
		fmt.Sprintf("Tempo totale: %.0f ms (%.1f s)", totalMs, totalMs/1000),
		fmt.Sprintf("Tempo medio: %.1f ms (+/- %.1f ms)", mean, std),
		fmt.Sprintf("Mediana: %.1f ms", median),
		fmt.Sprintf("Moda: %.1f ms", mode),
	}
	for _, s := range stats {
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="value">%s</text>`, svgPadding+20, y, s))
		y += rowHeight
	}
	y += sectionGap

	// Section 3: Accuracy chart
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="section-title">Accuracy nel Tempo</text>`, svgPadding, y))
	y += 25
	drawLineChart(&b, svgPadding, y, accuracy, "Accuracy %", 0, 100, "line-acc")
	y += chartHeight + sectionGap

	// Section 4: Netdata stats table
	if netdata != nil && len(netdata.Timestamps) > 0 {
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="section-title">Metriche di Sistema</text>`, svgPadding, y))
		y += 25

		netHeaders := []string{"Metrica", "Media", "StdDev", "Mediana", "Moda"}
		drawTableHeader(&b, svgPadding, y, netHeaders)
		y += rowHeight

		metrics := []struct {
			name string
			data []float64
		}{
			{"CPU", netdata.CPU},
			{"RAM", netdata.RAM},
			{"GPU", netdata.GPU},
			{"Disk", netdata.Disk},
		}
		for _, m := range metrics {
			if len(m.data) > 0 {
				drawTableRow(&b, svgPadding, y, []string{
					m.name,
					fmt.Sprintf("%.1f", Mean(m.data)),
					fmt.Sprintf("%.1f", StdDev(m.data)),
					fmt.Sprintf("%.1f", Median(m.data)),
					fmt.Sprintf("%.1f", Mode(m.data)),
				}, "")
				y += rowHeight
			}
		}
		y += sectionGap

		// Section 5: Netdata time chart
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="section-title">Metriche nel Tempo</text>`, svgPadding, y))
		y += 25
		drawMultiLineChart(&b, svgPadding, y, netdata)
		y += chartHeight + sectionGap
	}

	// Extras
	if len(eval.Extras) > 0 {
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="section-title">Extras</text>`, svgPadding, y))
		y += 25
		for _, ex := range eval.Extras {
			extraText := ex.Title
			if ex.Description != "" {
				extraText += " - " + ex.Description
			}
			if ex.Type == "text" {
				extraText += ": " + ex.Body
			} else if ex.Type == "image" {
				extraText += " [image: " + ex.Body + "]"
			}
			b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="value">%s</text>`, svgPadding+20, y, esc(extraText)))
			y += rowHeight
		}
	}

	b.WriteString(`</svg>`)
	return b.String()
}

func drawTableHeader(b *strings.Builder, x, y int, headers []string) {
	b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#f3f4f6"/>`, x, y-16, len(headers)*tableColWidth, rowHeight))
	for i, h := range headers {
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" font-weight="bold">%s</text>`, x+i*tableColWidth+10, y, h))
	}
}

func drawTableRow(b *strings.Builder, x, y int, cols []string, statusClass string) {
	for i, c := range cols {
		class := "value"
		if i == 1 && statusClass != "" {
			class = statusClass
		}
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="%s">%s</text>`, x+i*tableColWidth+10, y, class, esc(c)))
	}
}

func drawLineChart(b *strings.Builder, x, y int, data []float64, label string, minVal, maxVal float64, cssClass string) {
	w := svgWidth - 2*svgPadding
	h := chartHeight

	// Background
	b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#fafafa" stroke="#e5e7eb"/>`, x, y, w, h))

	// Grid lines
	for i := 0; i <= 4; i++ {
		gy := y + h - int(float64(i)*float64(h)/4)
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="grid"/>`, x, gy, x+w, gy))
		val := minVal + float64(i)*(maxVal-minVal)/4
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label">%.0f</text>`, x-45, gy+4, val))
	}

	if len(data) < 2 {
		return
	}

	// Draw line
	points := make([]string, len(data))
	for i, v := range data {
		px := x + int(float64(i)*float64(w)/float64(len(data)-1))
		py := y + h - int((v-minVal)/(maxVal-minVal)*float64(h))
		points[i] = fmt.Sprintf("%d,%d", px, py)
	}
	b.WriteString(fmt.Sprintf(`<polyline points="%s" class="%s"/>`, strings.Join(points, " "), cssClass))

	// Y axis label
	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" transform="rotate(-90,%d,%d)">%s</text>`, x-50, y+h/2, x-50, y+h/2, label))
}

func drawMultiLineChart(b *strings.Builder, x, y int, netdata *NetdataMetrics) {
	w := svgWidth - 2*svgPadding
	h := chartHeight

	b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#fafafa" stroke="#e5e7eb"/>`, x, y, w, h))

	// Find max value across all metrics
	maxVal := 0.0
	for _, v := range netdata.CPU {
		if v > maxVal {
			maxVal = v
		}
	}
	for _, v := range netdata.RAM {
		if v > maxVal {
			maxVal = v
		}
	}
	for _, v := range netdata.GPU {
		if v > maxVal {
			maxVal = v
		}
	}
	for _, v := range netdata.Disk {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 100
	}

	// Grid
	for i := 0; i <= 4; i++ {
		gy := y + h - int(float64(i)*float64(h)/4)
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="grid"/>`, x, gy, x+w, gy))
		val := float64(i) * maxVal / 4
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label">%.0f</text>`, x-45, gy+4, val))
	}

	n := len(netdata.Timestamps)
	if n < 2 {
		return
	}

	drawMetricLine(b, x, y, w, h, netdata.CPU, maxVal, n, "line-cpu")
	drawMetricLine(b, x, y, w, h, netdata.RAM, maxVal, n, "line-ram")
	drawMetricLine(b, x, y, w, h, netdata.GPU, maxVal, n, "line-gpu")
	drawMetricLine(b, x, y, w, h, netdata.Disk, maxVal, n, "line-disk")

	// Legend
	lx := x + w - 200
	ly := y + 20
	legend := []struct {
		label string
		color string
	}{
		{"CPU", "#3b82f6"},
		{"RAM", "#22c55e"},
		{"GPU", "#a855f7"},
		{"Disk", "#f97316"},
	}
	for i, l := range legend {
		b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="12" height="12" fill="%s"/>`, lx, ly+i*18-10, l.color))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label">%s</text>`, lx+18, ly+i*18, l.label))
	}
}

func drawMetricLine(b *strings.Builder, x, y, w, h int, data []float64, maxVal float64, n int, cssClass string) {
	points := make([]string, n)
	for i, v := range data {
		px := x + int(float64(i)*float64(w)/float64(n-1))
		py := y + h - int(v/maxVal*float64(h))
		points[i] = fmt.Sprintf("%d,%d", px, py)
	}
	b.WriteString(fmt.Sprintf(`<polyline points="%s" class="%s"/>`, strings.Join(points, " "), cssClass))
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
