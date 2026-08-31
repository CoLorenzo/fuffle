package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func generateHTML(eval *EvaluationFile, netdata *NetdataMetrics, hardware *HardwareInfo) string {
	durations := make([]float64, len(eval.Entries))
	for i, e := range eval.Entries {
		durations[i] = entryDuration(e)
	}

	mean := Mean(durations)
	std := StdDev(durations)
	startMin, endMax := totalDuration(eval.Entries)
	totalMs := float64(endMax - startMin)

	accuracy := calcAccuracyOverTime(eval.Entries)

	resultCounts := make(map[string]int)
	for _, e := range eval.Entries {
		r := e.Result
		if r == "" {
			r = "N/A"
		}
		resultCounts[r]++
	}

	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="it">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + escHTML(eval.Title) + `</title>
<style>
  :root {
    --bg: #f0f4f8;
    --card: #ffffff;
    --border: #3b82f6;
    --border-light: #e2e8f0;
    --text: #1e293b;
    --text-secondary: #64748b;
    --accent: #3b82f6;
    --ok: #22c55e;
    --fail: #ef4444;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
    background: var(--bg);
    color: var(--text);
    line-height: 1.6;
    padding: 2rem;
  }
  .page {
    max-width: 1000px;
    margin: 0 auto;
    background: var(--card);
    border: 2px solid var(--border);
    border-radius: 12px;
    padding: 2.5rem;
  }
  h1 {
    font-size: 2rem;
    text-align: center;
    margin-bottom: 0.25rem;
  }
  .subtitle {
    text-align: center;
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin-bottom: 2rem;
  }
  .section {
    border: 1px solid var(--border-light);
    border-radius: 12px;
    padding: 1.25rem 1.5rem;
    margin: 1.5rem 0;
  }
  .section-header h2 {
    font-size: 1.15rem;
    color: var(--accent);
  }
  .section-subtitle {
    font-size: 0.8rem;
    color: var(--text-secondary);
    margin-top: 0.15rem;
  }
  .card {
    background: var(--bg);
    border: 1px solid var(--border-light);
    border-radius: 10px;
    padding: 1.25rem;
    margin: 1rem 0;
  }
  .card h3 {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-secondary);
    margin-bottom: 0.75rem;
  }
  .results-summary {
    display: flex;
    gap: 2rem;
    align-items: center;
    flex-wrap: wrap;
  }
  .results-legend {
    flex: 1;
    min-width: 200px;
  }
  .results-legend h3 {
    font-size: 0.9rem;
    margin-bottom: 0.75rem;
    color: var(--text-secondary);
  }
  .legend-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.4rem;
    font-size: 0.85rem;
  }
  .legend-dot {
    width: 12px;
    height: 12px;
    border-radius: 3px;
    flex-shrink: 0;
  }
  .legend-label { flex: 1; }
  .legend-count { font-weight: bold; }
  .legend-pct { color: var(--text-secondary); font-size: 0.8rem; }
  .pie-container {
    flex-shrink: 0;
    position: relative;
    width: 150px;
    height: 150px;
  }
  .pie {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    mask: radial-gradient(closest-side, transparent 64%, #000 65%);
    -webkit-mask: radial-gradient(closest-side, transparent 64%, #000 65%);
    box-shadow: inset 0 0 0 1px var(--border-light);
  }
  .pie-center {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    font-size: 1.4rem;
    font-weight: bold;
  }
  .pie-center small {
    font-size: 0.6rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }
  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 1rem;
    margin: 1rem 0;
  }
  .stat-card {
    background: var(--bg);
    border: 1px solid var(--border-light);
    border-radius: 8px;
    padding: 1rem;
    text-align: center;
  }
  .stat-card .label {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-secondary);
    margin-bottom: 0.25rem;
  }
  .stat-card .value {
    font-size: 1.2rem;
    font-weight: bold;
  }
  .chart-container {
    overflow-x: auto;
  }
  .chart-container svg { display: block; }
  .chart {
    position: relative;
    height: 260px;
    overflow: hidden;
    border: 1px solid var(--border-light);
    border-radius: 8px;
    background: var(--bg);
  }
  .chart-body {
    position: absolute;
    top: 10px;
    bottom: 14px;
    left: 46px;
    right: 12px;
  }
  .chart-grid {
    position: absolute;
    inset: 0;
    background: repeating-linear-gradient(to top,
      transparent 0,
      transparent calc(25% - 1px),
      rgba(148, 163, 184, 0.35) calc(25% - 1px),
      rgba(148, 163, 184, 0.35) 25%);
  }
  .chart-fill { position: absolute; inset: 0; }
  .chart-line { position: absolute; inset: 0; }
  .chart-yaxis {
    position: absolute;
    top: 10px;
    bottom: 14px;
    left: 0;
    width: 40px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    font-size: 0.65rem;
    color: var(--text-secondary);
    text-align: right;
    line-height: 1;
  }
  .chart-xaxis {
    display: flex;
    justify-content: space-between;
    font-size: 0.7rem;
    color: var(--text-secondary);
    margin-top: 0.4rem;
  }
  .legend {
    display: flex;
    gap: 1.25rem;
    flex-wrap: wrap;
    margin-top: 0.75rem;
    font-size: 0.8rem;
  }
  .legend-item { display: flex; align-items: center; gap: 0.35rem; }
  .legend-dot { width: 10px; height: 10px; border-radius: 2px; display: inline-block; }
  .extras-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1rem;
    margin: 1rem 0;
  }
  .extra-card {
    background: var(--bg);
    border: 1px solid var(--border-light);
    border-radius: 8px;
    padding: 1.25rem;
  }
  .extra-card h3 { font-size: 1rem; margin-bottom: 0.3rem; }
  .extra-card .desc { font-size: 0.8rem; color: var(--text-secondary); margin-bottom: 0.5rem; }
  .extra-card .body { font-size: 0.85rem; line-height: 1.5; }
  .extra-card img { max-width: 100%; border-radius: 4px; margin-top: 0.5rem; }
  .results-link {
    display: block;
    text-align: center;
    margin-top: 2rem;
    padding: 0.75rem 1.5rem;
    background: var(--accent);
    color: white;
    text-decoration: none;
    border-radius: 8px;
    font-size: 0.9rem;
    font-weight: bold;
  }
  .results-link:hover { opacity: 0.9; }
  table { width: 100%; border-collapse: collapse; font-size: 0.85rem; margin-top: 1rem; }
  th {
    text-align: left;
    background: var(--bg);
    border-bottom: 2px solid var(--border-light);
    padding: 0.5rem 0.75rem;
    color: var(--text-secondary);
    font-weight: 600;
    text-transform: uppercase;
    font-size: 0.75rem;
  }
  td { padding: 0.45rem 0.75rem; border-bottom: 1px solid var(--border-light); }
  tr:hover td { background: var(--bg); }
  .ok { color: var(--ok); font-weight: bold; }
  .fail { color: var(--fail); font-weight: bold; }
  .back-link {
    display: inline-block;
    margin-bottom: 1rem;
    color: var(--accent);
    text-decoration: none;
    font-size: 0.9rem;
  }
</style>
</head>
<body>
<div class="page">
<h1>` + escHTML(eval.Title) + `</h1>
<p class="subtitle">Report generato</p>
`)

	total := len(eval.Entries)
	okCount := 0
	for _, e := range eval.Entries {
		if len(e.Tags) > 0 {
			okCount++
		}
	}

	// Section: Extra
	if len(eval.Extras) > 0 {
		b.WriteString(`<section class="section">
<div class="section-header">
<h2>Extra</h2>
<p class="section-subtitle">Contenuti aggiuntivi definiti dall'utente</p>
</div>
<div class="extras-grid">`)
		for _, ex := range eval.Extras {
			b.WriteString(buildExtraCard(ex))
		}
		b.WriteString(`</div>
</section>`)
	}

	// Section: Risultati
	b.WriteString(`<section class="section">
<div class="section-header">
<h2>Risultati</h2>
<p class="section-subtitle">Esito della valutazione sui file</p>
</div>`)

	// Card: Percentuali
	if total > 0 {
		colors := []string{"#22c55e", "#3b82f6", "#f59e0b", "#ef4444", "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16"}
		entries := make([]struct {
			Label string
			Count int
			Pct   float64
		}, 0, len(resultCounts))
		for label, count := range resultCounts {
			entries = append(entries, struct {
				Label string
				Count int
				Pct   float64
			}{label, count, float64(count) / float64(total) * 100})
		}
		for i := range entries {
			for j := i + 1; j < len(entries); j++ {
				if entries[j].Count > entries[i].Count {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		b.WriteString(`<div class="card">
<h3>Percentuali</h3>
<div class="results-summary">
<div class="results-legend">
<h3>Uniques (` + fmt.Sprintf("%d", len(entries)) + `)</h3>`)
		for i, e := range entries {
			c := colors[i%len(colors)]
			b.WriteString(fmt.Sprintf(`<div class="legend-row">
<span class="legend-dot" style="background:%s"></span>
<span class="legend-label">%s</span>
<span class="legend-count">%d</span>
<span class="legend-pct">(%s)</span>
</div>`, c, escHTML(e.Label), e.Count, fmt.Sprintf("%.1f%%", e.Pct)))
		}
		b.WriteString(`</div>
<div class="pie-container">
<div class="pie" style="background:conic-gradient(`)
		var cumPct float64
		for i, e := range entries {
			c := colors[i%len(colors)]
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(fmt.Sprintf(` %s %.2f%% %.2f%%`, c, cumPct, cumPct+e.Pct))
			cumPct += e.Pct
		}
		b.WriteString(`)"></div>
<div class="pie-center">` + fmt.Sprintf("%d", total) + `<small>file</small></div>
</div>
</div>`)
		b.WriteString(fmt.Sprintf(`<div class="stats">
  <div class="stat-card"><div class="label">File totali</div><div class="value">%d</div></div>
  <div class="stat-card"><div class="label">OK</div><div class="value" style="color:#22c55e">%d</div></div>
  <div class="stat-card"><div class="label">FAIL</div><div class="value" style="color:#ef4444">%d</div></div>
  <div class="stat-card"><div class="label">Tempo totale</div><div class="value">%.0f ms</div></div>
  <div class="stat-card"><div class="label">Tempo medio</div><div class="value">%.1f ms</div></div>
  <div class="stat-card"><div class="label">StdDev</div><div class="value">%.1f ms</div></div>
</div>`, len(eval.Entries), okCount, len(eval.Entries)-okCount, totalMs, mean, std))
		b.WriteString(`</div>`)
	}

	// Card: Accuracy nel tempo
	if len(accuracy) > 1 {
		b.WriteString(`<div class="card">
<h3>Accuracy nel Tempo</h3>
<div class="chart-container">`)
		b.WriteString(buildCSSAreaChart(accuracy, "Accuracy %", 0, 100, "#3b82f6", "Inizio", "Fine"))
		b.WriteString(`</div>
</div>`)
	}

	b.WriteString(`</section>`)

	// Section: Metriche
	hasNetdata := netdata != nil && len(netdata.Timestamps) > 0
	if hardware != nil || hasNetdata {
		b.WriteString(`<section class="section">
<div class="section-header">
<h2>Metriche</h2>
<p class="section-subtitle">Sistema e utilizzo delle risorse durante la valutazione</p>
</div>`)

		// Card: Sistema
		if hardware != nil {
			b.WriteString(`<div class="card">
<h3>Sistema</h3>
<div class="stats">`)

			if hardware.CPU.Model != "" {
				b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">CPU</div><div class="value">%s</div></div>`, escHTML(hardware.CPU.Model)))
				b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">Core / Thread</div><div class="value">%d / %d</div></div>`, hardware.CPU.Cores, hardware.CPU.Threads))
			}

			if len(hardware.GPU) > 0 {
				for _, gpu := range hardware.GPU {
					b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">GPU</div><div class="value">%s</div></div>`, escHTML(gpu.Model)))
					if gpu.VRAM != "" {
						b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">VRAM</div><div class="value">%s</div></div>`, escHTML(gpu.VRAM)))
					}
				}
			}

			if len(hardware.RAM) > 0 {
				for _, ram := range hardware.RAM {
					b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">RAM</div><div class="value">%s</div></div>`, escHTML(ram.Model)))
					b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">Clock / Tipo</div><div class="value">%s / %s</div></div>`, escHTML(ram.Clock), escHTML(ram.Technology)))
				}
			}

			if len(hardware.Disk) > 0 {
				for _, disk := range hardware.Disk {
					b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">Disco</div><div class="value">%s</div></div>`, escHTML(disk.Model)))
					b.WriteString(fmt.Sprintf(`<div class="stat-card"><div class="label">Tipo</div><div class="value">%s</div></div>`, escHTML(disk.Type)))
				}
			}

			b.WriteString(`</div>
</div>`)
		}

		// Card: Metriche nel tempo
		if hasNetdata {
			b.WriteString(`<div class="card">
<h3>Metriche nel Tempo</h3>
<div class="chart-container">`)
			b.WriteString(buildCSSMultiAreaChart(netdata))
			b.WriteString(`<div class="legend">
  <div class="legend-item"><span class="legend-dot" style="background:#3b82f6"></span> CPU</div>
  <div class="legend-item"><span class="legend-dot" style="background:#22c55e"></span> RAM</div>
  <div class="legend-item"><span class="legend-dot" style="background:#a855f7"></span> GPU</div>
  <div class="legend-item"><span class="legend-dot" style="background:#f97316"></span> Disk</div>
</div>`)
			b.WriteString(`</div>
</div>`)
		}

		b.WriteString(`</section>`)
	}

	// Link to results
	b.WriteString(`<a href="/results.html" class="results-link">Mostra Risultati Completi</a>`)

	b.WriteString(`</div></body></html>`)
	return b.String()
}

func generateResultsHTML(eval *EvaluationFile) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="it">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Risultati - ` + escHTML(eval.Title) + `</title>
<style>
  :root {
    --bg: #f0f4f8;
    --card: #ffffff;
    --border: #3b82f6;
    --border-light: #e2e8f0;
    --text: #1e293b;
    --text-secondary: #64748b;
    --accent: #3b82f6;
    --ok: #22c55e;
    --fail: #ef4444;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
    background: var(--bg);
    color: var(--text);
    line-height: 1.6;
    padding: 2rem;
  }
  .page {
    max-width: 1000px;
    margin: 0 auto;
    background: var(--card);
    border: 2px solid var(--border);
    border-radius: 12px;
    padding: 2.5rem;
  }
  h1 { font-size: 1.5rem; text-align: center; margin-bottom: 1.5rem; }
  .back-link {
    display: inline-block;
    margin-bottom: 1rem;
    color: var(--accent);
    text-decoration: none;
    font-size: 0.9rem;
  }
  table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
  th {
    text-align: left;
    background: var(--bg);
    border-bottom: 2px solid var(--border-light);
    padding: 0.5rem 0.75rem;
    color: var(--text-secondary);
    font-weight: 600;
    text-transform: uppercase;
    font-size: 0.75rem;
  }
  td { padding: 0.45rem 0.75rem; border-bottom: 1px solid var(--border-light); }
  tr:hover td { background: var(--bg); }
  .ok { color: var(--ok); font-weight: bold; }
  .fail { color: var(--fail); font-weight: bold; }
</style>
</head>
<body>
<div class="page">
<a href="/" class="back-link">← Torna al Report</a>
<h1>Risultati</h1>
<table>
<thead><tr><th>#</th><th>File</th><th>Stato</th><th>Tempo (ms)</th><th>Tag</th><th>Result</th></tr></thead>
<tbody>`)

	for i, e := range eval.Entries {
		status := "OK"
		cls := "ok"
		if len(e.Tags) == 0 {
			status = "FAIL"
			cls = "fail"
		}
		tag := ""
		if len(e.Tags) > 0 {
			tag = e.Tags[0]
		}
		b.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%s</td><td class=\"%s\">%s</td><td>%.0f</td><td>%s</td><td>%s</td></tr>",
			i+1, escHTML(e.File), cls, status, entryDuration(e), escHTML(tag), escHTML(e.Result)))
	}

	b.WriteString(`</tbody></table>
</div></body></html>`)
	return b.String()
}

const (
	htmlChartH     = 260.0
	htmlChartTop   = 10.0
	htmlChartBot   = 14.0
	htmlChartPadL  = 46.0
	htmlChartPadR  = 12.0
	htmlChartLineP = 2.0
)

func hexToRGBA(hex string, alpha float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return hex
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return fmt.Sprintf("rgba(%d,%d,%d,%.2f)", r, g, b, alpha)
}

func cssChartPolygons(data []float64, minVal, maxVal, lineWidthPct float64) (fillPts, linePts string) {
	n := len(data)
	if n == 0 {
		return "", ""
	}
	span := maxVal - minVal
	if span <= 0 {
		span = 1
	}
	x := func(i int) float64 {
		if n == 1 {
			return 50
		}
		return float64(i) / float64(n-1) * 100
	}
	y := func(v float64) float64 {
		return 100 - (v-minVal)/span*100
	}

	fill := make([]string, 0, n+2)
	fill = append(fill, "0% 100%")
	for i, v := range data {
		fill = append(fill, fmt.Sprintf("%.3f%% %.3f%%", x(i), y(v)))
	}
	fill = append(fill, "100% 100%")
	fillPts = strings.Join(fill, ", ")

	if lineWidthPct > 0 {
		line := make([]string, 0, 2*n)
		for i, v := range data {
			line = append(line, fmt.Sprintf("%.3f%% %.3f%%", x(i), y(v)-lineWidthPct))
		}
		for i := n - 1; i >= 0; i-- {
			line = append(line, fmt.Sprintf("%.3f%% %.3f%%", x(i), y(data[i])))
		}
		linePts = strings.Join(line, ", ")
	}
	return fillPts, linePts
}

func buildCSSAreaChart(data []float64, label string, minVal, maxVal float64, color, startLabel, endLabel string) string {
	if len(data) < 2 {
		return `<div class="chart-xaxis"><span>Dati insufficienti</span></div>`
	}
	t := htmlChartLineP / (htmlChartH - htmlChartTop - htmlChartBot) * 100
	fill, line := cssChartPolygons(data, minVal, maxVal, t)

	var b strings.Builder
	b.WriteString(`<div class="chart" title="` + label + `">`)
	b.WriteString(`<div class="chart-yaxis">`)
	for i := 0; i <= 4; i++ {
		val := minVal + float64(i)*(maxVal-minVal)/4
		b.WriteString(fmt.Sprintf(`<span>%.0f</span>`, val))
	}
	b.WriteString(`</div><div class="chart-body">`)
	b.WriteString(`<div class="chart-grid"></div>`)
	b.WriteString(fmt.Sprintf(`<div class="chart-fill" style="background:linear-gradient(to bottom,%s,%s);clip-path:polygon(%s)"></div>`,
		hexToRGBA(color, 0.20), hexToRGBA(color, 0.03), fill))
	b.WriteString(fmt.Sprintf(`<div class="chart-line" style="background:%s;clip-path:polygon(%s)"></div>`, color, line))
	b.WriteString(`</div></div>`)
	b.WriteString(fmt.Sprintf(`<div class="chart-xaxis"><span>%s</span><span>%s</span></div>`, startLabel, endLabel))
	return b.String()
}

func buildCSSMultiAreaChart(netdata *NetdataMetrics) string {
	n := len(netdata.Timestamps)
	if n < 2 {
		return `<div class="chart-xaxis"><span>Dati insufficienti</span></div>`
	}
	maxVal := 0.0
	for _, s := range [][]float64{netdata.CPU, netdata.RAM, netdata.GPU, netdata.Disk} {
		for _, v := range s {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal == 0 {
		maxVal = 100
	}

	series := []struct {
		name  string
		color string
		data  []float64
	}{
		{"CPU", "#3b82f6", netdata.CPU},
		{"RAM", "#22c55e", netdata.RAM},
		{"GPU", "#a855f7", netdata.GPU},
		{"Disk", "#f97316", netdata.Disk},
	}
	t := htmlChartLineP / (htmlChartH - htmlChartTop - htmlChartBot) * 100

	var b strings.Builder
	b.WriteString(`<div class="chart">`)
	b.WriteString(`<div class="chart-yaxis">`)
	for i := 0; i <= 4; i++ {
		b.WriteString(fmt.Sprintf(`<span>%.0f</span>`, float64(i)*maxVal/4))
	}
	b.WriteString(`</div><div class="chart-body">`)
	b.WriteString(`<div class="chart-grid"></div>`)
	for _, s := range series {
		fill, line := cssChartPolygons(s.data, 0, maxVal, t)
		if fill == "" {
			continue
		}
		b.WriteString(fmt.Sprintf(`<div class="chart-fill" title="%s" style="background:linear-gradient(to bottom,%s,%s);clip-path:polygon(%s)"></div>`,
			s.name, hexToRGBA(s.color, 0.14), hexToRGBA(s.color, 0.02), fill))
		b.WriteString(fmt.Sprintf(`<div class="chart-line" title="%s" style="background:%s;clip-path:polygon(%s)"></div>`,
			s.name, s.color, line))
	}
	b.WriteString(`</div></div>`)
	b.WriteString(fmt.Sprintf(`<div class="chart-xaxis"><span>%s</span><span>%s</span></div>`,
		formatTs(netdata.Timestamps[0]), formatTs(netdata.Timestamps[n-1])))
	return b.String()
}

func formatTs(ts int64) string {
	if ts > 1e11 {
		return time.UnixMilli(ts).Format("15:04:05")
	}
	return time.Unix(ts, 0).Format("15:04:05")
}

func buildExtraCard(ex Extra) string {
	var b strings.Builder
	b.WriteString(`<div class="extra-card">`)
	b.WriteString(`<h3>` + escHTML(ex.Title) + `</h3>`)
	if ex.Description != "" {
		b.WriteString(`<p class="desc">` + escHTML(ex.Description) + `</p>`)
	}
	b.WriteString(`<div class="body">`)
	switch ex.Type {
	case "image":
		b.WriteString(fmt.Sprintf(`<img src="/images/%s" alt="%s"/>`, escHTML(ex.Body), escHTML(ex.Title)))
	default:
		b.WriteString(`<p>` + escHTML(ex.Body) + `</p>`)
	}
	b.WriteString(`</div></div>`)
	return b.String()
}

func escHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
