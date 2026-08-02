# mautic-cli

Mautic REST API のコマンドラインクライアント。バイナリ名は `mautic`。

複数の Mautic インスタンスを**名前付きプロファイル**で切り替えられ、**Basic 認証と
OAuth2 の両方**に対応し、Mautic API のほぼ全エンドポイントを noun→verb の
サブコマンドとして持つ。ラップしていないエンドポイントも `mautic api` から叩ける。

AI エージェントから使うための知識は `mautic llm` がバイナリ内から出力する
（→ [AI エージェントから使う](#ai-エージェントから使う)）。

```console
$ mautic auth status --verify
instance:      https://mautic.example.com
api endpoint:  https://mautic.example.com/api
auth method:   basic
identity:      api-user
source:        env

verification:  OK — authenticated as api-user <ops@example.com>, role Administrator
mautic version: 7.0.1

$ mautic contacts list --where email:isNotNull --limit 5
id     fields.all.email        fields.all.firstname  points  lastActive
1042   taro@example.com        太郎                    12      2026-08-01T10:42:14+00:00
1043   hanako@example.jp       花子                    3       2026-08-01T09:11:02+00:00
showing 5 of 431 — use --limit or --all for more
```

## インストール

```bash
# GitHub Releases から
VERSION=$(curl -fsSL https://api.github.com/repos/ideamans/mautic-cli/releases/latest \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64; [ "$ARCH" = "aarch64" ] && ARCH=arm64
curl -fsSL "https://github.com/ideamans/mautic-cli/releases/download/${VERSION}/mautic-cli_${VERSION#v}_${OS}_${ARCH}.tar.gz" \
  | tar -xz -C /tmp
mv /tmp/mautic ~/.local/bin/

# ソースから
go install github.com/ideamans/mautic-cli/cmd/mautic@latest
```

アーカイブ名は `mautic-cli_…`（goreleaser の `project_name`）だが、**中身の
バイナリは `mautic`**。

## 認証

Mautic は **Basic 認証**（Mautic ユーザーの username / password。Mautic 側の
設定 → API 設定で有効化が必要）と **OAuth2** の両方をサポートする。スクリプトや
エージェントからは Basic 認証が簡単。

解決順序は **フラグ > 環境変数 > 設定ファイル**。環境変数が揃っていれば
プロファイルより優先される —— これが「プロファイルが5個あるラップトップ」と
「何も設定されていない CI ランナー」を同じコードパスで動かす仕組み。

### 環境変数だけで使う

```bash
export MAUTIC_BASE_URL=https://mautic.example.com   # 末尾の /api は付いていてもよい
export MAUTIC_API_USER=api-user
export MAUTIC_API_KEY=s3cret
mautic contacts list
```

| 変数 | 意味 | 別名 |
| --- | --- | --- |
| `MAUTIC_BASE_URL` | インスタンスの URL | `MAUTIC_API_ENDPOINT`, `MAUTIC_URL` |
| `MAUTIC_API_USER` | Basic 認証のユーザー名 | `MAUTIC_USERNAME`, `MAUTIC_USER` |
| `MAUTIC_API_KEY` | Basic 認証のパスワード | `MAUTIC_PASSWORD` |
| `MAUTIC_CLIENT_ID` / `MAUTIC_CLIENT_SECRET` | OAuth2 クライアント | — |
| `MAUTIC_ACCESS_TOKEN` | OAuth2 アクセストークン | — |
| `MAUTIC_PROFILE` | 使用するプロファイル | — |

### プロファイルで複数インスタンスを切り替える

```bash
mautic configure --name prod  --url https://mautic.example.com --username api-user --password s3cret
mautic configure --name stage --url https://stage.example.com  --username api-user --password other

mautic profiles list
mautic profiles use prod
mautic contacts list --profile stage      # 1回だけ切り替える
```

設定は `~/.config/mautic-cli/`（`XDG_CONFIG_HOME` を尊重）に置かれ、
非機密の `config.toml` と機密の `credentials.json`（0600）に分かれる。

### OAuth2

Mautic 側で 設定 → API クレデンシャル から OAuth2 クライアントを作り、
**リダイレクト URI を完全一致で登録**する（既定は
`http://localhost:8080/oauth/callback`）。

```bash
mautic auth login --url https://mautic.example.com \
  --client-id 1_abc --client-secret def
```

ブラウザが開いて認可すると、トークンがプロファイルに保存される。以後は
アクセストークンが期限切れになると自動でリフレッシュされ、**ローテーションされた
リフレッシュトークンも保存し直される**（Mautic はリフレッシュのたびに新しい
リフレッシュトークンを発行するため、これを取りこぼすと再認可が必要になる）。

ブラウザを使えない環境では `--client-credentials` を使う（インスタンスが
有効にしていれば動く）。

## 使い方

```
mautic <リソース> <動詞> [引数] [フラグ]
```

```bash
mautic contacts list --search 'email:someone@example.com'
mautic contacts get 1234
mautic contacts create --email a@example.com --firstname A
mautic contacts activity 1234 --limit 20
mautic segments add-contact 7 1234
mautic emails send-to-contact 42 1234 --dry-run
mautic api GET /stats/page_hits --query limit=5      # ラップしていないものは何でも
```

対応リソース: assets, campaigns, categories, companies, contacts,
dynamic-contents, emails, fields, files, focus, forms, messages, notes,
notifications, pages, points, point-triggers, reports, roles, segments, smses,
stages, stats, tags, themes, tweets, users, webhooks, data。

### 出力

既定は人間向けのテーブル（セルは48桁で切り詰め、列は絞り込み済み）。
**パースするときは必ず `--format json` / `jsonl` / `csv`** を使う。

```bash
mautic contacts list --format json | jq '.total, .count'
mautic contacts list --fields id,fields.all.email,points --format csv
mautic stats query page_hits --where date_hit:gte:2026-08-01 --all --format jsonl
```

`--format json` のリストは常に `{"total":N,"count":M,"items":[…]}`。
`total` はコレクション全体、`count` はこの呼び出しが返した件数。

### 書き込みの安全装置

- `--dry-run` — 送信するリクエストを表示して、**送らない**
- `--yes` — 破壊的コマンドの確認ゲートを解除する。付けなければ実行されない
- 書き込み時は `target: <インスタンス> as <ユーザー>` を stderr に出す

```bash
mautic emails send 42 --dry-run    # 何が起きるか見る
mautic emails send 42 --yes        # 実行する
```

## AI エージェントから使う

```bash
mautic llm                  # 全リファレンス（Markdown）
mautic llm --format json    # 章ごとの JSON 配列
```

`mautic llm` はバイナリに埋め込まれているので、**実行中のバイナリと必ず一致する**。
オフラインで完結し、ドキュメントサイトのように古くならない。章立ては
ガイド（鉄則・認証・質問→コマンド表・失敗モード）、落とし穴、分析、
コマンドカタログ（生成物）。

カタログは長いので、必要な章だけ読む:

```bash
mautic llm | sed -n '1,/^# Pitfalls/p'
mautic llm --format json | jq -r '.[] | select(.file=="20-analytics.md") | .body'
mautic contacts list --help
```

Claude Code などから使う場合は、プラグイン `mautic-cli` をインストールすると
`/mautic-usage` と `/mautic-install` のスキルが入る。

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install mautic-cli@ideamans-plugins
```

Copilot / Cursor / Gemini CLI 等では同じスキルを `gh skill` で入れられる。

```bash
gh skill install ideamans/mautic-cli/plugins/mautic-cli/skills/mautic-usage --agent copilot
```

## 開発

```bash
go build ./...
go generate ./...     # internal/llmdocs/90-commands.md を再生成（生成物）
go test ./...
```

詳細は [CLAUDE.md](CLAUDE.md)。

## ライセンス

MIT
