---
title: zipp
description: Command-line zip utilities — list, timestamp and update zip archives.
---

## Usage

Zip utilities:

- `zipls`: lists zip contents
- `zipts`: creates a zip file with a timestamp suffix in the name
- `zipw`: add file to existing zip

## Zipw

Add file to zip.

Add the file `README.md` to the zip `dist/zipp-0.7.0-dev_linux_amd64.zip` in the path `zipp-0.7.0-dev_linux_amd64/myreadme.md`:

```bash
$ zipw -f README.md -i 'zipp-0.7.0-dev_linux_amd64/myreadme.md' -z dist/zipp-0.7.0-dev_linux_amd64.zip
$ zipls dist/zipp-0.7.0-dev_linux_amd64.zip
zipp-0.7.0-dev_linux_amd64/myreadme.md
zipp-0.7.0-dev_linux_amd64/zipls
zipp-0.7.0-dev_linux_amd64/zipw
zipp-0.7.0-dev_linux_amd64/zipts
```

Nested write is supported.

Add the file `README.md` in the path `com/example/readme.md` into the file `WEB-INF/lib/library.jar` into the file `webapp.war` into the file `corporate.ear`:

```bash
$ zipw -f README.md -i 'webapp.war#WEB-INF/lib/library.jar#com/example/readme.md' -z corporate.ear
```

Parameters can also be passed via a YAML file:

```bash
$ zipw -p zipw.yml
```

If neither `-f` nor `-p` is provided, `zipw` enters interactive mode and prompts for the parameters.

Preview what would happen without modifying any file:

```bash
$ zipw -f README.md -i 'archive/readme.md' -z app.zip --dry-run
```

Available flags:

| Flag | Description |
|------|-------------|
| `-f, --file`     | File to add to the archive |
| `-z, --zip`      | Zip archive to update |
| `-i, --inner`    | Inner path inside the archive (defaults to the source file path) |
| `-p, --params`   | YAML file containing parameters |
| `-N, --dry-run`  | Print what would happen without modifying any file |
| `--force`        | Overwrite existing entries without prompting |
| `--no-overwrite` | Abort with an error if the entry already exists |
| `-q, --quiet`    | Quiet mode |
| `-V, --verbose`  | Verbose mode |

## Zipls

List contents of a zip file.

```bash
$ zipls dist/zipp-0.7.0-dev_linux_amd64.zip
NAME                                  TYPE
--------------------------------------
zipp-0.7.0-dev_linux_amd64/          dir
zipp-0.7.0-dev_linux_amd64/zipls     file
zipp-0.7.0-dev_linux_amd64/zipw      file
zipp-0.7.0-dev_linux_amd64/zipts     file
```

Show file sizes and modification times:

```bash
$ zipls -s -t dist/zipp-0.7.0-dev_linux_amd64.zip
```

Filter entries by pattern (highlighted in output):

```bash
$ zipls --grep zipw dist/zipp-0.7.0-dev_linux_amd64.zip
```

Exclude entries matching a pattern:

```bash
$ zipls --exclude test dist/zipp-0.7.0-dev_linux_amd64.zip
```

Sort results:

```bash
$ zipls --sort size dist/zipp-0.7.0-dev_linux_amd64.zip
$ zipls --sort date --reverse dist/zipp-0.7.0-dev_linux_amd64.zip
```

Available flags:

| Flag | Description |
|------|-------------|
| `-g, --grep`    | Filter entries containing pattern (highlighted in output) |
| `-e, --exclude` | Exclude entries containing pattern |
| `-d, --dir`     | Show only directories |
| `-s, --size`    | Show file size |
| `-t, --time`    | Show modification time |
| `--sort`        | Sort order: `none` (default), `name`, `size`, `date` |
| `-r, --reverse` | Reverse sort order |

## Zipts

Create a zip file with timestamp suffix:

```bash
$ zipts testdata/
Zipping  /home/enrico/Projects/zipp/testdata
Completed /home/enrico/Projects/zipp/testdata-20210227173843.zip
```

It is possible to exclude files from zip:

```bash
$ zipts -x '\.git/*' -x 'vendor/*' .
```

Write the archive to a specific output directory:

```bash
$ zipts -o /tmp/backups .
```

Preview the output path without creating the archive:

```bash
$ zipts -N .
```

Available flags:

| Flag | Description |
|------|-------------|
| `-x, --exclude` | Exclude files matching pattern (repeatable) |
| `-o, --out`     | Output directory (defaults to parent of input path) |
| `-N, --noop`    | Print output path without creating the archive |
| `-q, --quiet`   | Quiet mode |
| `-V, --verbose` | Verbose mode |

## License

[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0) — Copyright (C) 2016-TODAY zipp contributors.
