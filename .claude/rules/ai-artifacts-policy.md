# 派生物と原本の対応

このリポジトリには「生成されるファイル」と「手で書くファイル」が混在している。
生成物を直接編集しても次の `go generate ./...` で消える。原本を直す。

| 派生物 | 原本 | 再生成 |
| --- | --- | --- |
| `internal/llmdocs/90-commands.md` | `cmd/` 以下の cobra コマンド定義（`Use` / `Short` / `Long` / `Example` / フラグ説明） | `go generate ./...` |

生成物は **git 追跡する**。`go:embed` がビルド時に実ファイルを要求するため
ignore できない。代わりに CI が `go generate ./...` → `git diff --exit-code` で
ズレを検出する。

## 手で書くもの（生成しない）

| ファイル | 内容 |
| --- | --- |
| `internal/llmdocs/00-guide.md` | 鉄則・認証・質問→コマンド表・ワークフロー・失敗モード |
| `internal/llmdocs/10-pitfalls.md` | エラーにならず間違った答えになる罠 |
| `internal/llmdocs/20-analytics.md` | stats テーブルと data データセットの使い分け |
| `plugins/mautic-cli/skills/*/SKILL.md` | 配布スキル。**英語**、frontmatter は5フィールドのみ |
| `context7.json` の `rules` | 落とし穴のみ。**英語** |
| `README.md` / `CLAUDE.md` | 人間向け。日本語でよい |

SKILL.md と `context7.json` の rules を生成しないのは意図的な判断。どちらも
「読み物」であり、どの落とし穴を書くかの取捨選択は人間の仕事だから。

## バージョンの原本

`cmd/root.go` の `const PluginVersion` が単一の原本。

- `plugins/mautic-cli/.claude-plugin/plugin.json` の `version` ↔ `PluginVersion`
  … `cmd/mautic/plugin_test.go` が検査する
- `PluginVersion` ↔ git タグ … `.github/workflows/release.yml` が検査する

`--version` の出力をパースして比較しないこと。出力書式は自分で決めたテンプレート
であり、変えた瞬間にリリースが壊れる。
