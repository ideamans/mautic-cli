---
name: regen-ai
description: mautic-cli の AI 向け派生物を再生成し、検証して結果を報告する。コマンド定義やフラグ、配布スキルを変更したあとに実行する。
allowed-tools: Bash Read
---

# /regen-ai

`internal/llmdocs/90-commands.md` を cobra のコマンドツリーから作り直し、
検証まで通す。

## 手順

```bash
cd /Users/miyanaga/dev/mautic-cli

go generate ./...          # 90-commands.md を再生成
go build ./...
go test ./...              # TestPluginSkills とバージョン整合を含む
go run ./cmd/mautic llm | head -5
go run ./cmd/mautic llm --format json | head -5
git status --short internal/llmdocs
```

## 報告すること

1. `90-commands.md` に差分が出たか。出たなら、どのコマンド／フラグが増減したか
   （`git diff --stat` と、`git diff` の追加行から見出しを拾う）
2. `go test ./...` の結果
3. **手書き側の更新漏れの指摘**。ここが本来の価値。生成物は自動で追随するが、
   運用知識は追随しない。差分を見て、次のどれかに当たるなら指摘する:
   - コマンドやフラグが増えた → `00-guide.md` の「Which command answers which
     question」表に載せるべきか
   - 破壊的な操作が増えた → 鉄則6と `mautic-usage` SKILL.md の
     「2. Before writing anything」の表
   - 出力形式・既定値が変わった → 鉄則と Failure modes
   - 新しい罠が判明した → `10-pitfalls.md` と `context7.json` の `rules`

差分が無ければ「派生物は最新」と一行で報告して終わる。

## やらないこと

- `90-commands.md` を手で編集しない。コマンド定義を直して再生成する
- `00-guide.md` を勝手に書き換えない。**指摘までが仕事**で、内容の追加は
  ユーザーの確認を取ってから
