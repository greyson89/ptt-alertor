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
