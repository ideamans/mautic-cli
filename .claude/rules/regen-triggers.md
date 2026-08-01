---
paths:
  - cmd/**/*.go
  - internal/llmdocs/[0-8]*.md
  - plugins/mautic-cli/**
---

# SSOT を触ったら再生成する

いま編集したファイルは、AIナレッジの原本のいずれかです。

## コマンド定義（`cmd/**/*.go`）を変えた

`internal/llmdocs/90-commands.md` が古くなっています。

```bash
go generate ./...
go test ./...
```

差分をコミットに含めること。CI が `git diff --exit-code` で落とします。

さらに、以下のどれかに当てはまるなら **`00-guide.md` も手で更新**してください。
生成物は追随しますが、運用知識は追随しません。

- コマンドやフラグを**足した** → 「質問→コマンド」表に1行足すか、既存の行を直す
- **既定値・出力形式・終了コードを変えた** → 鉄則と失敗モードの表
- 破壊的な操作を足した → 鉄則6と、`plugins/mautic-cli/skills/mautic-usage/SKILL.md`
  の「2. Before writing anything」の表
- エージェントが間違えやすい罠を見つけた → `10-pitfalls.md` と `context7.json`
  の `rules`

## 手書き章（`internal/llmdocs/0*.md`〜`8*.md`）を変えた

再生成は不要です。ただし `mautic llm` が壊れていないか確認してください。

```bash
go run ./cmd/mautic llm | head -5
go run ./cmd/mautic llm --format json | head -5
```

`00-guide.md` は**約100行**に保つこと。40行を超えるトピックは
`10-` 〜 `89-` の独立した章に切り出します。エージェントが全部読む入口だからです。

## 配布スキル（`plugins/mautic-cli/**`）を変えた

```bash
go test ./cmd/mautic/
```

`TestPluginSkills` が次を検査します。

- `plugin.json.version` == `cmd/root.go` の `PluginVersion`
- frontmatter が5フィールド（`name` / `description` / `license` /
  `compatibility` / `allowed-tools`）のみ。`argument-hint` などの
  Claude 専用フィールドは他エージェントに無視されるため**書かない**
- discovery キーワードが description に残っていること
- `-install` スキルが存在すること

SKILL.md は**英語**で書きます。他社エージェントがこの description で
道具を発見するためです。
