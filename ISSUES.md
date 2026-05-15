# Zipp — Issue Backlog

> Generato il 2026-05-15. Nessuna issue aperta su GitHub.

---

## [FEAT-01] `zipls`: output in formato JSON e CSV

**Tipo**: Feature  
**Componente**: `zipls`  
**Priorità**: Media

### Descrizione

`zipls` produce attualmente solo output testuale non strutturato, incompatibile con pipeline di scripting o tool come `jq`. Aggiungere un flag `--format` che permetta di scegliere il formato di output.

### Comportamento atteso

```sh
zipls archive.zip --format json
zipls archive.zip --format csv
```

Output JSON (array di oggetti):

```json
[
  { "name": "docs/index.adoc", "size": 12680, "size_human": "12.4 kB", "is_dir": false, "modified": "2024-03-01T10:00:00Z" },
  { "name": "cmd/", "size": 0, "size_human": "0 B", "is_dir": true, "modified": "2024-02-28T08:30:00Z" }
]
```

Output CSV (con header):

```
name,size,size_human,is_dir,modified
docs/index.adoc,12680,12.4 kB,false,2024-03-01T10:00:00Z
cmd/,0,0 B,true,2024-02-28T08:30:00Z
```

### Criteri di accettazione

- [ ] Flag `--format` accetta i valori `text` (default), `json`, `csv`
- [ ] Con `--format json` l'output è JSON valido, stampato su stdout
- [ ] Con `--format csv` la prima riga è l'header, le righe successive i dati
- [ ] I flag `-g`, `-e`, `-d`, `-s` restano compatibili con tutti i formati
- [ ] Con `--format json` o `--format csv`, i messaggi di errore vanno su stderr (non inquinano stdout)
- [ ] Aggiornare help e documentazione

---

## [FEAT-02] Completamento shell (Bash, Zsh, Fish)

**Tipo**: Feature  
**Componente**: `zipls`, `zipts`, `zipw`  
**Priorità**: Bassa

### Descrizione

Nessuno dei tre comandi supporta l'autocompletamento shell. L'utente deve digitare flag e path manualmente, rendendo la CLI meno ergonomica rispetto a tool moderni. `urfave/cli` (già in uso) offre supporto nativo per il completamento — va solo attivato e documentato.

### Comportamento atteso

```sh
# Bash
zipls --<TAB>          # mostra: --size --grep --exclude --dir --format --sort --help --version
zipls arch<TAB>        # completa il path del file .zip

# Zsh / Fish — stesso comportamento
```

Script di installazione:

```sh
zipls --generate-bash-completion >> ~/.bash_completion
zipts --generate-bash-completion >> ~/.bash_completion
zipw  --generate-bash-completion >> ~/.bash_completion
```

### Criteri di accettazione

- [ ] Abilitare `EnableBashCompletion` in tutti e tre i comandi via `urfave/cli`
- [ ] Il completamento suggerisce tutti i flag definiti
- [ ] Il completamento suggerisce file `.zip` per gli argomenti posizionali
- [ ] Aggiungere sezione "Shell completion" al README e a `docs/index.adoc`
- [ ] Testare su Bash ≥ 4, Zsh ≥ 5, Fish ≥ 3

---

## [UX-01] `zipls`: output come tabella con colonne allineate

**Tipo**: Miglioramento UX  
**Componente**: `zipls`  
**Priorità**: Alta

### Descrizione

L'output attuale di `zipls` è testo semplice non allineato. Su archivi con molte entry o nomi lunghi diventa difficile da leggere. Strutturare l'output come tabella con colonne fisse migliora significativamente la leggibilità senza richiedere tool esterni.

### Comportamento atteso

```
NAME                          SIZE        MODIFIED             TYPE
cmd/zipls/main.go             4.1 kB      2024-02-28 08:30     file
cmd/zipts/main.go             5.3 kB      2024-03-01 10:00     file
docs/                         —           2024-01-15 12:00     dir
docs/index.adoc               12.4 kB     2024-03-01 10:00     file
```

- La colonna `SIZE` appare solo se `-s` è attivo (comportamento attuale preservato)
- La colonna `MODIFIED` appare solo se `-t` / `--time` è attivo (nuovo flag)
- La larghezza delle colonne si adatta dinamicamente al contenuto
- I match di `-g` restano evidenziati in verde nella colonna `NAME`

### Criteri di accettazione

- [ ] L'output di default usa il formato tabella allineata
- [ ] La larghezza di ogni colonna si calcola sul massimo dell'intero set prima di stampare
- [ ] Aggiungere flag `-t` / `--time` per mostrare la colonna data di modifica
- [ ] Con `--format json` o `--format csv` (FEAT-01) il formato tabella non si applica
- [ ] Con `--no-color` o `NO_COLOR` impostato, nessun codice ANSI nell'output
- [ ] La regressione su output piped (`zipls archive.zip | grep foo`) non introduce caratteri spuri

---

## [UX-02] Barra di progresso per `zipts` e `zipw`

**Tipo**: Miglioramento UX  
**Componente**: `zipts`, `zipw`  
**Priorità**: Media

### Descrizione

Operando su archivi o directory di grandi dimensioni, `zipts` e `zipw` non forniscono alcun feedback visivo durante l'elaborazione. L'utente non può distinguere tra un processo lento e uno bloccato. Una barra di progresso risolve questo problema.

### Comportamento atteso

```
Compressing project/ ...
  [=================>        ] 68%  142/208 files  4.2 MB / 6.1 MB
Done: project-20260515143022.zip (6.1 MB) in 3.2s
```

Regole di visualizzazione:

- La barra appare solo se stdout è un TTY (non in pipe o redirect)
- Con `-q` / `--quiet` la barra è soppressa
- Con `-V` / `--verbose` la barra è soppressa e al suo posto appaiono i nomi file uno per riga
- La barra si aggiorna in-place (no flood di righe)

### Criteri di accettazione

- [ ] Implementare barra di progresso per `zipts` durante la compressione
- [ ] Implementare barra di progresso per `zipw` durante aggiunta e re-archiviazione di ZIP annidati
- [ ] La barra mostra: percentuale, numero file processati/totali, dimensione processata/totale
- [ ] Rilevare TTY (`isatty` già in dipendenza) e non mostrare la barra in ambiente non interattivo
- [ ] `-q` sopprime la barra; `-V` la sostituisce con log per file
- [ ] Aggiungere test che verificano l'assenza di output barra in modalità non-TTY

---

## [UX-03] `zipw`: modalità dry-run (`--dry-run` / `-N`)

**Tipo**: Miglioramento UX  
**Componente**: `zipw`  
**Priorità**: Alta

### Descrizione

`zipts` supporta già `-N` / `--noop` per simulare l'operazione senza scrivere su disco. `zipw` non ha questa funzionalità, ma ne ha ancora più bisogno: operando su ZIP annidati critici (es. `.ear`, `.war`) un'operazione errata può corrompere l'archivio. Il dry-run permette di verificare il comportamento prima di agire.

### Comportamento atteso

```sh
zipw -N -i com/example/Config.class library.jar config/Config.class
```

Output atteso (nessun file modificato):

```
[dry-run] Would add: config/Config.class → library.jar :: com/example/Config.class
[dry-run] No files written.
```

Per ZIP annidati:

```sh
zipw -N -i com/example/Config.class corporate.ear#webapp.war#library.jar config/Config.class
```

```
[dry-run] Would extract: corporate.ear → /tmp/zipp-dry-xxxxxx/
[dry-run] Would extract: webapp.war → /tmp/zipp-dry-xxxxxx/webapp/
[dry-run] Would add: config/Config.class → library.jar :: com/example/Config.class
[dry-run] Would re-pack: library.jar, webapp.war, corporate.ear
[dry-run] No files written.
```

### Criteri di accettazione

- [ ] Aggiungere flag `-N` / `--dry-run` a `zipw`, coerente con `zipts`
- [ ] In modalità dry-run, nessun file viene creato, modificato o eliminato
- [ ] L'output descrive ogni operazione che sarebbe eseguita, incluse le estrazioni temporanee per ZIP annidati
- [ ] Il messaggio finale conferma esplicitamente che nessun file è stato scritto
- [ ] Exit code 0 se la simulazione non rileva errori prevedibili, non-zero altrimenti
- [ ] Aggiornare help, README e `docs/index.adoc`

---

## [UX-04] Messaggi di errore contestuali con suggerimenti

**Tipo**: Miglioramento UX  
**Componente**: `zipls`, `zipts`, `zipw`  
**Priorità**: Media

### Descrizione

Gli errori attuali sono brevi e tecnici (es. `"file not found"`, exit code numerico). Non guidano l'utente verso la soluzione. Aggiungere una riga `Hint:` contestuale a ogni categoria di errore riduce il tempo di debug e migliora l'esperienza dei nuovi utenti.

### Mappatura errori → suggerimenti

| Situazione | Messaggio attuale | Nuovo hint |
|---|---|---|
| File ZIP non trovato | `file not found` | `Hint: use 'zipts .' to create a new archive from the current directory` |
| Argomento mancante | `missing argument` | `Hint: run 'zipls --help' to see usage and examples` |
| File non è un ZIP valido | `invalid zip file` | `Hint: verify the file is a valid ZIP archive with 'file <name>'` |
| Path interno non trovato in `zipw` | `entry not found` | `Hint: use 'zipls <archive>' to list available entries` |
| Directory output non esiste (`zipts`) | `output dir not found` | `Hint: create the directory first with 'mkdir -p <path>'` |
| ZIP annidato non trovato (`zipw` con `#`) | `nested zip not found` | `Hint: check the nesting path with 'zipls <outer.zip>'` |

### Formato output

```
Error: file "archive.zip" not found
Hint:  use 'zipts .' to create a new archive from the current directory
```

- `Error:` in rosso (se colori abilitati)
- `Hint:` in giallo (se colori abilitati)
- Entrambi su stderr

### Criteri di accettazione

- [ ] Definire una mappa centralizzata `errorCode → hint string` in `lib/core`
- [ ] Ogni categoria di errore esistente ha un hint associato
- [ ] I hint usano i colori ANSI se il terminale è interattivo, testo plain altrimenti
- [ ] I hint rispettano `NO_COLOR` e `-q`
- [ ] Con `-q` i hint sono soppressi (solo il messaggio di errore essenziale)
- [ ] Aggiornare i test E2E per verificare la presenza dei messaggi hint sullo stderr corretto

---

## [UX-05] `zipls`: ordinamento dell'output

**Tipo**: Miglioramento UX  
**Componente**: `zipls`  
**Priorità**: Media

### Descrizione

`zipls` elenca le entry nell'ordine in cui sono memorizzate nell'archivio ZIP. Su archivi grandi, trovare i file più grandi o più recenti richiede di fare pipe con `sort`. Aggiungere un flag `--sort` nativo elimina questa frizione.

### Comportamento atteso

```sh
zipls archive.zip --sort size     # dal più grande al più piccolo
zipls archive.zip --sort name     # ordine alfabetico crescente
zipls archive.zip --sort date     # dal più recente al più vecchio
zipls archive.zip --sort size --reverse   # dal più piccolo al più grande
```

### Valori accettati per `--sort`

| Valore | Criterio |
|--------|----------|
| `none` | Ordine interno dell'archivio (default attuale) |
| `name` | Alfabetico sul path completo |
| `size` | Dimensione non compressa, decrescente |
| `date` | Data di modifica, decrescente |

### Criteri di accettazione

- [ ] Aggiungere flag `--sort` con i valori `none`, `name`, `size`, `date`
- [ ] Aggiungere flag `--reverse` / `-r` per invertire l'ordinamento
- [ ] Il default di `--sort` è `none` (nessuna regressione sul comportamento attuale)
- [ ] `--sort size` richiede la lettura della dimensione indipendentemente dal flag `-s`
- [ ] L'ordinamento è compatibile con `-g`, `-e`, `-d` (si applica dopo il filtraggio)
- [ ] Aggiornare help e documentazione

---

## [UX-06] `zipw`: conferma interattiva prima di sovrascrivere entry esistenti

**Tipo**: Miglioramento UX  
**Componente**: `zipw`  
**Priorità**: Alta

### Descrizione

`zipw` sovrascrive silenziosamente le entry già presenti nell'archivio senza alcun avviso. Su archivi di produzione (`.ear`, `.war`, `.jar`) questo può causare perdita irreversibile di dati. Aggiungere un prompt di conferma interattivo, disattivabile esplicitamente con `--force`.

### Comportamento atteso

Se l'entry esiste già nell'archivio:

```
Warning: 'com/example/Config.class' already exists in library.jar
Overwrite? [y/N]: _
```

- Default: **No** (invio senza digitare = non sovrascrive, exit 0)
- `y` o `Y`: sovrascrive e procede
- `n`, `N` o invio: salta la entry e prosegue (se batch) o esce (se singola)

Flag aggiuntivi:

```sh
zipw --force ...     # sovrascrive sempre senza prompt
zipw --no-overwrite ...   # fallisce con errore se la entry esiste già (exit non-zero)
```

### Criteri di accettazione

- [ ] Prima di sovrascrivere una entry esistente, mostrare il prompt su stderr
- [ ] Il prompt appare solo se stdin è un TTY; in pipe/script si comporta come `--no-overwrite` (errore sicuro)
- [ ] `--force` disabilita il prompt e sovrascrive sempre
- [ ] `--no-overwrite` disabilita il prompt e fallisce con exit non-zero se la entry esiste
- [ ] Con `-q` in ambiente non-TTY il comportamento di default è `--no-overwrite`
- [ ] Per operazioni batch (YAML params), il prompt appare per ogni conflitto o si applica `--force` / `--no-overwrite` globalmente
- [ ] Aggiornare i test E2E per coprire i casi: overwrite confermato, overwrite rifiutato, `--force`, `--no-overwrite`
- [ ] Aggiornare help e documentazione
