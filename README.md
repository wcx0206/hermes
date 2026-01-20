# Hermes

> **A lightweight Cloud Backup Automator powered by Rclone.**

`Hermes`, named after the messenger of the Greek gods, is a robust toolset designed to automate your cloud backup workflow with ease.

## 📖 Overview

The project consists of two main components:

- **`hermes-backup` (Daemon)**: A background process that executes scheduled backup tasks via `rclone` based on `config.yaml`.
- **`hermes` (CLI)**: A command-line tool for managing projects, editing configurations, controlling the service, and triggering manual backups.

## 📦 Installation

1. **Via Script**:

   ```bash
   curl -fsSL https://hermes.elysiapro.cn/install.sh | bash
   ```

   The script downloads `hermes-backup` and `hermes` to `~/.local/bin`.
2. **Manual**: Download binaries directly from GitHub Releases.

---

## 🚀 Usage Guide

### 1. Configuration Management

- **Flag Logic**: `--config` defaults to `config.yaml` in the same directory as the executable.
- **Show Config**: `hermes config show [--config path]`.
- **Edit Config**: `hermes config edit [--config path]`.

### 2. Project Management

- **Flag Logic**: `--config` default follows the logic mentioned above.
- **List Projects**: `hermes project list` (or `ls`).
- **Operations**:
  - `hermes project add [--config path]`: Add a new backup project.
  - `hermes project update <name> [--config path]`: Update an existing project.
  - `hermes project delete <name> [--config path]`: Remove a project.

### 3. Server Control (Daemon)

- **Flag Logic**:
  - `--binary`: Defaults to `hermes-backup` in the same directory as the CLI tool.
  - `--config`: Default logic same as above.
- **Commands**:
  - `hermes server start [--config path] [--binary path]`: Launch the background daemon.
  - `hermes server stop [--config path] [--binary path]`: Stop the background daemon.
  - `hermes server restart [--config path] [--binary path]`: Restart the background daemon.

### 4. Manual Backup Trigger

- **Immediate Run**: `hermes backup run --projects <name1,name2>`.
- **Note**: Project names must be **comma-separated**.

---

## 📄 Configuration Example (`config.yaml`)

YAML

```yaml
logging:
  path: /hermes/logs/service.log # Log file path
  debug: true # Whether to print logs to stdout

defaults:
  rclone_remote: aliyun # Default rclone remote name
  bucket: racknerd-vps # Default bucket name
  cron: 0 1 * * * # Default schedule
  mode: copy # Default mode: copy/sync

projects:
  - name: vaultwarden
    mode: sync
    source_paths:
      - /opt/vaultwarden/data
    cron: 0 2 * * *
    rclone_remotes:
      - name: aliyun
        bucket: racknerd-vps
```
