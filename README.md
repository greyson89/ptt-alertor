# Ptt-Alertor

<img align="right" src="https://raw.githubusercontent.com/watain666/ptt-alertor/master/logo.jpg">

[![Build Status](https://github.com/watain666/ptt-alertor/actions/workflows/main.yml/badge.svg)](https://github.com/watain666/ptt-alertor/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/watain666/ptt-alertor/branch/master/graph/badge.svg)](https://codecov.io/gh/watain666/ptt-alertor)
[![Go Report Card](https://goreportcard.com/badge/github.com/watain666/ptt-alertor)](https://goreportcard.com/report/github.com/watain666/ptt-alertor)
[![Code Climate](https://api.codeclimate.com/v1/badges/f7047295fce56a0465dc/maintainability)](https://codeclimate.com/github/watain666/ptt-alertor/maintainability)
[![StackShare](https://img.shields.io/badge/tech-stack-0690fa.svg?style=flat)](https://stackshare.io/watain666/ptt-alertor)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Docker Deployment

Ptt-Alertor persists all of its data (users, subscriptions, boards, articles) in a
single SQLite file — no Redis, DynamoDB, or Postgres required.

1. (Optional) copy env template if you want to customize settings:

```bash
cp .env.example .env
```

2. Update `.env` with your Telegram bot token (see [@BotFather](https://t.me/BotFather))
   and other settings.

3. Start the app:

```bash
docker compose up --build -d
```

The SQLite database is stored on the `sqlite-data` named volume (`/storage/ptt-alertor.db`
inside the container), so it survives container rebuilds. To back it up, copy that file out
of the volume; to reset all data, remove the volume.

4. Open app:

- http://localhost:9090

5. Stop services:

```bash
docker compose down
```

## Native Deployment (no Docker, e.g. Windows)

Since storage is a single SQLite file with a pure-Go driver (no CGO), the binary has
no external runtime dependencies and runs directly on Windows/macOS/Linux.

1. Build (Go 1.26+):

```bash
go build -o ptt-alertor .          # or ptt-alertor.exe on Windows
```

2. The binary looks for its `public/` (templates) and `storage/` (SQLite file)
   folders next to itself if they aren't found in the current working
   directory — so it's safe to launch it via a shortcut, Task Scheduler, or a
   Windows service, regardless of what directory that launcher starts in.

3. Set the environment variables it needs (there is no `.env` auto-loading
   outside Docker Compose) and run it. On Windows, copy `run.ps1.example` to
   `run.ps1`, fill in your settings, and run `.\run.ps1`.

4. Telegram connection mode is chosen automatically from `APP_HOST`:
   - `APP_HOST` starts with `https://` → **webhook** mode: Telegram pushes
     updates to `POST /telegram/:token`. This needs a publicly reachable
     HTTPS URL — a real domain, or a tunnel such as
     [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)
     or [ngrok](https://ngrok.com/).
   - Otherwise (e.g. `http://localhost:9090`, the default for local/native
     runs) → **long polling** mode: the app itself repeatedly asks Telegram
     for new updates. No public URL, port forwarding, or TLS certificate
     needed — this is what you want when running on your own PC with no
     domain, and it's the default for exactly that reason.

5. To keep it running in the background on Windows, wrap it as a service with
   [NSSM](https://nssm.cc/) or schedule it to start at login via Task Scheduler.

## API

### Board

- GET /boards

- GET /boards/[board name]/articles

- GET /boards/[board name]/articles/[article code]

### Keyword

- GET /keyword/boards

### Author

- GET /author/boards

### PushSum

- GET /pushsum/boards

### Articles

- GET /articles

### User (Auth)

- GET /users

- GET /users/[account]

- POST /users

```json
{
  "profile": {
    "account": "sample",
    "telegram": "sample",
    "telegramChat": 123456789
  },
  "subscribes": [
    {
      "board": "gossiping",
      "keywords": ["問卦", "爆卦", "公告"]
    },
    {
      "board": "lol",
      "keywords": ["閒聊"]
    }
  ]
}
```

- PUT /users/[account]

```json
{
  "profile": {
    "account": "sample",
    "telegram": "sample",
    "telegramChat": 123456789
  },
  "subscribes": []
}
```

## Credits

### Real Life

Rose Li, Aries Huang, Scott Kao, Amy Li

### Ptt

DMM, oas, bestpika, Zero0910, lucky0509, wbreeze, chang0206, lindo0130, hungys, gyman7788, tooilxui, myamyakoko, whkuo, papago89, timeline, Kamikiri

### Facebook

Mr.clu, Woqeker
