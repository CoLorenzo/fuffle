# Piano: Gestione Sessioni

## Obiettivo
Aggiungere un sistema di sessioni per costruire file evaluation.yaml in modo incrementale, poi utilizzabili con `--report`.

## Comandi

```bash
# Crea una nuova sessione
fuffle session new

# Inserisce un risultato nella sessione
fuffle session insert -r "result" -f file.py --start 123 --end 456 --tags ok
```

## Formato session.yaml

```yaml
title: ""
entries:
  - file: "test.py"
    start_date: 1693123456789
    end_date: 1693123457890
    tags: ["ok"]
    result: "successo"
```

Identico a EvaluationEntry + campo `result` (testo libero).

## Modifiche

### 1. Modificare: `evaluation.go`
- Aggiungere campo `Result string` a `EvaluationEntry` con tag YAML `result`

### 2. Nuovo file: `session.go`
- `SessionFile` struct (identica a `EvaluationFile` ma con title opzionale)
- `sessionNew()` → crea `session.yaml` vuoto
- `sessionInsert(args)` → parsing flag `-r`, `-f`, `--start`, `--end`, `--tags`, legge session.yaml, aggiunge entry, riscrive

### 3. Modificare: `main.go`
- Aggiungere case `"session"` nello switch di `main()`
- `runSession(args)` → dispatch a `sessionNew()` o `sessionInsert()`
- Aggiungere a `printUsage()`

## Note
- `session.yaml` e' backward compatible con `--report` (campo `result` ignorato se non presente)
- `session new` sovrascrive se esiste gia' (con warning)
- `session new` NON chiede conferma se il file esiste gia' (comportamento semplice)
