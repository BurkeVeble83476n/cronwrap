# cronwrap

A drop-in cron job wrapper that adds structured logging, alerting, and execution history to any shell command.

---

## Installation

```bash
go install github.com/yourname/cronwrap@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwrap.git && cd cronwrap && go build -o cronwrap .
```

---

## Usage

Wrap any existing cron command by prepending `cronwrap`:

```bash
# Before
0 2 * * * /usr/local/bin/backup.sh

# After
0 2 * * * cronwrap --job backup --alert-on-failure run /usr/local/bin/backup.sh
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--job` | Name used to identify the job in logs and history |
| `--alert-on-failure` | Send an alert if the command exits with a non-zero status |
| `--timeout` | Maximum allowed runtime (e.g. `--timeout 30m`) |
| `--log-file` | Path to write structured JSON logs (default: stdout) |

**View execution history:**

```bash
cronwrap history --job backup
```

Output includes start time, duration, exit code, and captured stdout/stderr for each run.

---

## Features

- Structured JSON logging for every execution
- Alerting via webhook, email, or PagerDuty on failure or timeout
- Persistent execution history stored locally
- Zero changes required to your existing scripts

---

## License

MIT © yourname