# CLAUDE.md — mautic-cli

Mautic REST API の CLI。**リポジトリ名 `mautic-cli` / バイナリ名 `mautic` /
goreleaser の `project_name` は `mautic-cli`** で、バイナリだけ名前が違う。
アーカイブの中身は `mautic`、`go install .../cmd/mautic@latest` も `mautic`。

**public リポジトリ**。エージェントへの配布は `mautic llm`・context7・
public marketplace・`gh skill` の4経路。

## 変更時の必須手順

**機能を追加した、フラグを増やした、既存の挙動を変えた — このいずれかをしたら、
3か所すべてを更新してから終わること。**

| 更新先 | 対象 | やり方 |
| --- | --- | --- |
| ① ドキュメント | `README.md` | 使い方が変わったときのみ |
| ② ヘルプ | cobra の `Short` / `Long` / `Example` / フラグ説明 | コード内。**カタログはここから生成される** |
| ③ **LLMナレッジ** | `internal/llmdocs/00-guide.md` | 鉄則・認証・質問→コマンド表・失敗モードが変わったら（手書き、約100行を保つ） |
| | `internal/llmdocs/10-pitfalls.md` | エラーにならず間違った答えになる罠を見つけたら（手書き） |
| | `internal/llmdocs/20-analytics.md` | stats / data の使い分けが変わったら（手書き） |
| | `internal/llmdocs/90-commands.md` | **生成物。手編集しない** → `go generate ./...` |
| | `plugins/mautic-cli/skills/*/SKILL.md` | 手順や前提が変わったとき（**英語**） |
| | `context7.json` の `rules` | 落とし穴が増えたとき（**英語**） |

③ を忘れやすい。ドキュメントとヘルプは人間が読んで気づくが、**LLMナレッジが
古いことには誰も気づかない**（エージェントが黙って間違えるだけで、症状は
「なぜかうまく使えない」という形でしか現れない）。

判断の目安:

- コマンドやフラグを足した → ②を書いてから `go generate ./...`。使い方が非自明
  なら `00-guide.md` の「Which command answers which question」表にも1行
- 既定値・出力形式・終了コードを変えた → ①②③すべて
- **破壊的操作を追加した** → 鉄則6（`00-guide.md`）と `mautic-usage` SKILL.md の
  「2. Before writing anything」の表。`confirmed()` ゲートと `--dry-run` も忘れずに
- エージェントが間違えやすい罠を見つけた → `10-pitfalls.md` と `context7.json`
  の `rules`

## このリポジトリ固有の設計判断

**コマンドツリーはテーブル駆動**。`cmd/registry.go` の `resources` に1エントリ
足せば list/get/create/update/delete が生える。CRUD に収まらない操作は同じ
エントリの `Extra` フックに書く（`cmd/extras.go` / `cmd/contacts.go`）。
`/fields` `/files` `/themes` `/stats` `/data` だけはテーブルで表現できないので
`cmd/bespoke.go` に手書き。

**エンベロープキーは推測できない**。`/segments` は `{"lists": […]}`、
`/users/self` はラップすらされない。`registry.go` のコメントに一覧がある。
新しいリソースを足すときは公式ドキュメントで確認すること。

**グローバルフラグを影に隠すローカルフラグを作らない**。認証情報はどのコマンド
でも上書きできる必要がある（CI がそれで動く）。`webhooks --webhook-url`、
`users --login` / `--user-password`、`forms --form-fields` はそのための改名。
`TestNoCommandShadowsAGlobalFlag` が強制する。

**`--dry-run` は非 GET のみ抑止する**。`smses send` は Mautic 側が GET なので
抑止されない。これは `10-pitfalls.md` に書いてある。

**`coerce()` は先頭ゼロを文字列のまま残す**。`--set zipcode=0120` が 120 に
なると無言でレコードが壊れる。日本の郵便番号・電話番号で実際に踏む。

## 確認

```bash
go generate ./...     # コマンドカタログを作り直す
git diff --exit-code  # 差分が出たらコミット漏れ
go test ./...         # SKILL.md 検証とバージョン整合を含む
go run ./cmd/mautic llm | head
goreleaser check
```

実インスタンスに対する動作確認は、環境変数 `MAUTIC_API_ENDPOINT` /
`MAUTIC_API_USER` / `MAUTIC_API_KEY` が設定されていれば読み取り系はそのまま
通る（`mautic auth status --verify`）。**書き込み系を本番インスタンスで試さない
こと。** 試すなら必ず `--dry-run` から。

## リリース

`PluginVersion`（`cmd/root.go`）と `plugin.json` の `version` と git タグの3つを
揃える。テストとリリースワークフローが不一致を検出する。手順は
`plugins/mautic-cli/PUBLISH.md`。

## 参照

- 標準: <https://github.com/ideamans/go-llm-cli-kit/blob/main/LLM.md>
- CLI の作り方の標準: `ai-friendly-cli` スキル
- Mautic API: <https://github.com/mautic/developer-documentation>（`source/includes/_api_endpoint_*.md` が一次情報）
- 生成物と原本の対応: `.claude/rules/ai-artifacts-policy.md`
- 再生成: `/regen-ai`
