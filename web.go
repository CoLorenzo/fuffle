package main

import (
	"fmt"
	"strings"
)

func generateHTML(eval *EvaluationFile, netdata *NetdataMetrics, hardware *HardwareInfo) string {
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

	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="it">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + escHTML(eval.Title) + `</title>
<style>
  :root {
    --bg: #ffffff;
    --surface: #f8fafc;
    --border: #e2e8f0;
    --text: #1e293b;
    --text-secondary: #64748b;
    --accent: #3b82f6;
    --ok: #22c55e;
    --fail: #ef4444;
    --cpu: #3b82f6;
    --ram: #22c55e;
    --gpu: #a855f7;
    --disk: #f97316;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
    background: var(--bg);
    color: var(--text);
    line-height: 1.6;
    padding: 2rem;
  }
  .container { max-width: 1100px; margin: 0 auto; }
  h1 { font-size: 1.8rem; margin-bottom: 0.25rem; }
  .subtitle { color: var(--text-secondary); font-size: 0.9rem; margin-bottom: 2rem; }
  h2 {
    font-size: 1.2rem;
    color: var(--accent);
    border-bottom: 2px solid var(--border);
    padding-bottom: 0.4rem;
    margin: 2rem 0 1rem 0;
  }
  table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
  th {
    text-align: left;
    background: var(--surface);
    border-bottom: 2px solid var(--border);
    padding: 0.5rem 0.75rem;
    color: var(--text-secondary);
    font-weight: 600;
    text-transform: uppercase;
    font-size: 0.75rem;
    letter-spacing: 0.05em;
  }
  td { padding: 0.45rem 0.75rem; border-bottom: 1px solid var(--border); }
  tr:hover td { background: var(--surface); }
  .ok { color: var(--ok); font-weight: bold; }
  .fail { color: var(--fail); font-weight: bold; }
  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 1rem;
    margin: 1rem 0;
  }
  .stat-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1rem 1.25rem;
  }
  .stat-card .label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-secondary);
    margin-bottom: 0.25rem;
  }
  .stat-card .value { font-size: 1.4rem; font-weight: bold; }
  .chart-container {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1rem;
    margin: 1rem 0;
    overflow-x: auto;
  }
  .chart-container svg { display: block; }
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
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 1rem;
    margin: 1rem 0;
  }
  .extra-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.25rem;
  }
  .extra-card h3 {
    font-size: 1rem;
    margin-bottom: 0.3rem;
    color: var(--text);
  }
  .extra-card .desc {
    font-size: 0.8rem;
    color: var(--text-secondary);
    margin-bottom: 0.75rem;
  }
  .extra-card .body { font-size: 0.85rem; line-height: 1.5; }
  .extra-card img {
    max-width: 100%;
    border-radius: 4px;
    margin-top: 0.5rem;
  }
</style>
</head>
<body>
<div class="container">
<h1>` + escHTML(eval.Title) + `</h1>
<p class="subtitle">Report generato</p>
`)

	// Results table
	b.WriteString(`<h2>Risultati</h2>
<table>
<thead><tr><th>File</th><th>Stato</th><th>Tempo (ms)</th><th>Tag</th></tr></thead>
<tbody>`)
	for _, e := range eval.Entries {
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
		b.WriteString(fmt.Sprintf("<tr><td>%s</td><td class=\"%s\">%s</td><td>%.0f</td><td>%s</td></tr>",
			escHTML(e.File), cls, status, entryDuration(e), escHTML(tag)))
	}
	b.WriteString(`</tbody></table>`)

	// Stats cards
	b.WriteString(fmt.Sprintf(`<h2>Statistiche Tempo</h2>
<div class="stats">
  <div class="stat-card"><div class="label">File totali</div><div class="value">%d</div></div>
  <div class="stat-card"><div class="label">Tempo totale</div><div class="value">%.0f ms</div></div>
  <div class="stat-card"><div class="label">Tempo medio</div><div class="value">%.1f ms</div></div>
  <div class="stat-card"><div class="label">StdDev</div><div class="value">%.1f ms</div></div>
  <div class="stat-card"><div class="label">Mediana</div><div class="value">%.1f ms</div></div>
  <div class="stat-card"><div class="label">Moda</div><div class="value">%.1f ms</div></div>
</div>`, len(eval.Entries), totalMs, mean, std, median, mode))

	// Accuracy chart
	b.WriteString(`<h2>Accuracy nel Tempo</h2>
<div class="chart-container">`)
	b.WriteString(buildHTMLLineChart(accuracy, "Accuracy %", 0, 100, "#3b82f6"))
	b.WriteString(`</div>`)

	// Netdata
	if netdata != nil && len(netdata.Timestamps) > 0 {
		b.WriteString(`<h2>Metriche di Sistema</h2>
<table>
<thead><tr><th>Metrica</th><th>Media</th><th>StdDev</th><th>Mediana</th><th>Moda</th></tr></thead>
<tbody>`)
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
				b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%.1f</td><td>%.1f</td><td>%.1f</td><td>%.1f</td></tr>",
					m.name, Mean(m.data), StdDev(m.data), Median(m.data), Mode(m.data)))
			}
		}
		b.WriteString(`</tbody></table>`)

		b.WriteString(`<h2>Metriche nel Tempo</h2>
<div class="chart-container">`)
		b.WriteString(buildHTMLMultiLineChart(netdata))
		b.WriteString(`<div class="legend">
  <div class="legend-item"><span class="legend-dot" style="background:#3b82f6"></span> CPU</div>
  <div class="legend-item"><span class="legend-dot" style="background:#22c55e"></span> RAM</div>
  <div class="legend-item"><span class="legend-dot" style="background:#a855f7"></span> GPU</div>
  <div class="legend-item"><span class="legend-dot" style="background:#f97316"></span> Disk</div>
</div>`)
		b.WriteString(`</div>`)
	}

	// Hardware info
	if hardware != nil {
		b.WriteString(`<h2>Sistema</h2>
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

		b.WriteString(`</div>`)
	}

	// Extras
	if len(eval.Extras) > 0 {
		b.WriteString(`<h2>Extras</h2>
<div class="extras-grid">`)
		for _, ex := range eval.Extras {
			b.WriteString(buildExtraCard(ex))
		}
		b.WriteString(`</div>`)
	}

	b.WriteString(`</div></body></html>`)
	return b.String()
}

const htmlChartW = 1000
const htmlChartH = 250
const htmlChartPad = 50

func buildHTMLLineChart(data []float64, label string, minVal, maxVal float64, color string) string {
	var b strings.Builder
	w := htmlChartW
	h := htmlChartH

	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h+40, w, h+40))

	for i := 0; i <= 4; i++ {
		gy := 20 + h - i*h/4
		val := minVal + float64(i)*(maxVal-minVal)/4
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#e2e8f0" stroke-width="1"/>`, htmlChartPad, gy, w, gy))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="11" fill="#94a3b8" font-family="monospace">%.0f</text>`, htmlChartPad-40, gy+4, val))
	}

	if len(data) < 2 {
		b.WriteString(`</svg>`)
		return b.String()
	}

	points := make([]string, len(data))
	for i, v := range data {
		px := htmlChartPad + i*(w-htmlChartPad)/(len(data)-1)
		py := 20 + h - int((v-minVal)/(maxVal-minVal)*float64(h))
		points[i] = fmt.Sprintf("%d,%d", px, py)
	}
	b.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="2"/>`, strings.Join(points, " "), color))

	b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="11" fill="#94a3b8" font-family="monospace" transform="rotate(-90,%d,%d)">%s</text>`, htmlChartPad-45, 20+h/2, htmlChartPad-45, 20+h/2, label))

	b.WriteString(`</svg>`)
	return b.String()
}

func buildHTMLMultiLineChart(netdata *NetdataMetrics) string {
	var b strings.Builder
	w := htmlChartW
	h := htmlChartH

	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h+40, w, h+40))

	maxVal := 0.0
	for _, sets := range [][]float64{netdata.CPU, netdata.RAM, netdata.GPU, netdata.Disk} {
		for _, v := range sets {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal == 0 {
		maxVal = 100
	}

	for i := 0; i <= 4; i++ {
		gy := 20 + h - i*h/4
		val := float64(i) * maxVal / 4
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#e2e8f0" stroke-width="1"/>`, htmlChartPad, gy, w, gy))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="11" fill="#94a3b8" font-family="monospace">%.0f</text>`, htmlChartPad-40, gy+4, val))
	}

	n := len(netdata.Timestamps)
	if n < 2 {
		b.WriteString(`</svg>`)
		return b.String()
	}

	type series struct {
		data  []float64
		color string
	}
	allSeries := []series{
		{netdata.CPU, "#3b82f6"},
		{netdata.RAM, "#22c55e"},
		{netdata.GPU, "#a855f7"},
		{netdata.Disk, "#f97316"},
	}

	for _, s := range allSeries {
		points := make([]string, n)
		for i, v := range s.data {
			px := htmlChartPad + i*(w-htmlChartPad)/(n-1)
			py := 20 + h - int(v/maxVal*float64(h))
			points[i] = fmt.Sprintf("%d,%d", px, py)
		}
		b.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="2"/>`, strings.Join(points, " "), s.color))
	}

	b.WriteString(`</svg>`)
	return b.String()
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
