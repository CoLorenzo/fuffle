# Piano: Convertire report SVG in pagina web con --serve

## Obiettivo
Invece di generare un file SVG statico, generare una pagina HTML con CSS moderno e servirla con un server HTTP integrato.

## Formato YAML aggiornato

```yaml
title: "Nome Report"
netdatafile: "metrics.csv"
entries:
  - file: "test.py"
    start_date: 1234567890
    end_date: 1234567900
    tags: ["ok"]
extras:
  - type: text
    title: "Nota"
    description: "Spiegazione breve"
    body: "Testo del body qui"
  - type: image
    title: "Grafico"
    description: "Immagine di supporto"
    body: "./img/report.png"
```

- `title` resta top-level (stringa semplice)
- `extras` diventa `[]Extra` con `type`, `title`, `description`, `body`
- `body` per `type: image` è un path locale (relativo al CWD)

## Modifiche

### 1. Modificare: `evaluation.go`
- Nuovo struct `Extra`: `Type`, `Title`, `Description`, `Body` (con tag YAML)
- `EvaluationFile.Extras` cambia da `[]string` a `[]Extra`

### 2. Nuovo file: `web.go`
- `generateHTML(eval, netdata)` → genera stringa HTML completa con CSS embedded
- `buildSVGLineChart()` e `buildSVGMultiLineChart()` per grafici SVG inline nell'HTML
- Layout responsive, light mode, monospace font
- Cards per gli extras: `type: text` → card con testo, `type: image` → card con `<img>` (il path viene servito dal server)
- `buildExtraCard(extra)` genera la HTML per ogni extra

### 3. Nuovo file: `serve.go`
- `serveHTML(htmlContent, addr)` → avvia `http.ListenAndServe`
- Route `/` serve l'HTML
- Route `/images/*` serve i file immagine dal CWD (per extras image)
- `openBrowser(url)` → esegue `xdg-open` (Linux) o `open` (macOS)

### 4. Modificare: `main.go`
- `--report` accetta `--serve [:port]` opzionale
- Se `--serve`: genera HTML, chiama `serveHTML()`, apre browser
- Se `--serve` assente + `--output` termina con `.html`: genera HTML statico
- Se `--serve` assente + `--output` termina con `.svg` o nessun `--output`: SVG come prima (legacy)
- Aggiornare `printUsage()` con documentazione `--serve`

## Output atteso
```bash
fuffle --report eval.yaml --output report.svg    # SVG (invariato)
fuffle --report eval.yaml --output report.html   # HTML statico
fuffle --report eval.yaml --serve :8080          # Serve + apri browser
fuffle --report eval.yaml --serve                # Serve su :8080 + apri browser
```

## Nessuna nuova dependency
Solo stdlib Go: `net/http`, `os/exec`, `path/filepath`
