# Piano: Sessione completa

## Obiettivo
`fuffle session new` crea un file `session.yaml` completo con title, bigfetchfile, netdatafile, extras - pronto per essere usato con `--report`.

## Modifiche

### 1. `session.go` - Aggiornare SessionFile
```go
type SessionFile struct {
    Title        string            `yaml:"title,omitempty"`
    BigfetchFile string            `yaml:"bigfetchfile,omitempty"`
    NetdataFile  string            `yaml:"netdatafile,omitempty"`
    Extras       []Extra           `yaml:"extras,omitempty"`
    Info         SessionInfo       `yaml:"info,omitempty"`
    Entries      []EvaluationEntry `yaml:"entries"`
}
```

### 2. `session.go` - Aggiornare sessionNew
- Chiede title (input interattivo)
- Chiede bigfetchfile (opzionale, default "")
- Chiede netdatafile (opzionale, default "")
- Crea file YAML

### 3. Output atteso
```yaml
title: "Nome Report"
bigfetchfile: "bigfetch.json"
netdatafile: "metrics.json"
extras: []
info:
    starttime: 0
    endtime: 0
entries: []
```

### 4. Uso
```bash
fuffle session new  # interattivo
fuffle session insert -f test.py --start 123 --end 456 --tags ok
fuffle --report session.yaml --serve
```
