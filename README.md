# pipesnap

Lightweight utility to snapshot and replay stdin streams for debugging long-running pipeline processes.

---

## Installation

```bash
go install github.com/yourusername/pipesnap@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/pipesnap.git && cd pipesnap && go build -o pipesnap .
```

---

## Usage

**Snapshot** a stream to a file while passing it through:

```bash
some-long-running-process | pipesnap snap --out snapshot.bin | downstream-process
```

**Replay** a saved snapshot into any command:

```bash
pipesnap replay --in snapshot.bin | downstream-process
```

**Inspect** snapshot metadata:

```bash
pipesnap info --in snapshot.bin
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--out` | Output file for snapshot | `snap.bin` |
| `--in` | Input snapshot file to replay | — |
| `--compress` | Compress snapshot with gzip | `false` |
| `--timestamp` | Preserve inter-chunk timing on replay | `false` |

---

## Example

Debug a flaky ETL pipeline by capturing the exact input that caused a failure:

```bash
kafka-consumer | pipesnap snap --out debug.bin | etl-processor
# later, reproduce the issue:
pipesnap replay --in debug.bin | etl-processor
```

---

## License

MIT © [yourusername](https://github.com/yourusername)