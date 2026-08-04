# Publishing the mautic-cli plugin

このリポジトリは **public** です。プラグインは `ideamans/claude-public-plugins`
に登録し、context7 にも登録します。

## Before every release

1. `go generate ./...` — `internal/llmdocs/90-commands.md` を再生成し、差分が
   出たらコミットする。
2. `go test ./...` — `TestPluginSkills` を含む。`plugin.json.version` が
   `cmd/root.go` の `PluginVersion` と一致すること、SKILL.md の frontmatter が
   Agent Skills 標準（5フィールドのみ）に収まっていることを強制する。
3. `claude plugin validate plugins/mautic-cli`
4. `goreleaser check`
5. `PluginVersion` と `plugin.json.version` を同じコミットで上げる。リリース
   ワークフローがタグとの不一致を検出して止まる。

## Tagging

タグは2種類ある。**両方打つ。**

```bash
git checkout main && git pull

# ① バイナリのリリース。release ワークフローが v* で発火し goreleaser が走る
git tag -a v0.2.0 -m "v0.2.0"
git push origin v0.2.0

# ② プラグインのリリースマーカー。plugin.json と marketplace エントリの
#    整合を検証したうえで mautic-cli--v0.2.0 を打つ。v* にマッチしないので
#    リリースワークフローは発火しない
claude plugin tag plugins/mautic-cli --push -m "mautic-cli %s"
```

marketplace のエントリは `git-subdir` でデフォルトブランチを追うため、②を
打っても配信内容は変わらない（`plugin.json` の version がそのまま配られる）。
②はバージョンの記録と、`plugin.json` と marketplace エントリの食い違いを
リリース時に検出するためのもの。

## Registering in the marketplace (first release only)

`ideamans/claude-public-plugins`（marketplace 名 `ideamans-plugins`。gridgram と
chartjs2img が既に入っている）の `.claude-plugin/marketplace.json` の
`plugins` 配列に追加する。既存エントリと同じ形に揃えること:

```json
{
  "name": "mautic-cli",
  "source": {
    "source": "git-subdir",
    "url": "https://github.com/ideamans/mautic-cli.git",
    "path": "plugins/mautic-cli"
  },
  "description": "Adds /mautic-usage and /mautic-install skills for Mautic marketing automation — contacts, segments, campaigns, emails and the stats tables from the command line.",
  "homepage": "https://github.com/ideamans/mautic-cli",
  "repository": "https://github.com/ideamans/mautic-cli",
  "license": "MIT",
  "keywords": ["mautic", "marketing-automation", "crm", "email-marketing", "cli", "llm"],
  "category": "productivity"
}
```

`git-subdir` は本体のデフォルトブランチを指すため、以後の SKILL.md 更新は
マーケットプレイス側を触らずに配信される。裏返しに、破壊的変更も即座に届く。

`marketplace-register` スキルを使うこと。

## Registering with context7 (public repos only)

タグを push して public になってから、最後に実行する。

```bash
curl -sS -X POST https://context7.com/api/v2/add/repo/github \
  -H "Authorization: Bearer $CONTEXT7_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"docsRepoUrl":"https://github.com/ideamans/mautic-cli"}'
```

インデックスは数分かかる。`state: finalized` になるまで待つ:

```bash
curl -sS -G https://context7.com/api/v2/search --data-urlencode 'query=mautic-cli' \
  -H "Authorization: Bearer $CONTEXT7_API_KEY" \
  | jq -r '.results[] | select(.id=="/ideamans/mautic-cli") | .state'
```

インデックス中のリポジトリは検索結果に**現れない**（404 も返る）。これはエラー
ではないので、最初の 404 を失敗と判断しないこと。

### ⚠️ 未解決: llmdocs が取り込まれていない（2026-08-04 時点）

初回インデックスで `internal/llmdocs/` が1チャンクも入らず、ルートの
`README.md` だけになっている（`totalTokens=3135`）。`context7.json` の
`folders` には指定済みなので指定漏れではない。詳細と仮説は
リポジトリ直下の `CLAUDE.md`。

**`POST /api/v1/refresh` には10日のクールダウンがあり、最終更新
2026-08-02 に対して 2026-08-12 以降でないと再実行できない。**
次にこのファイルを読んだのがその日以降なら、まず以下を実行すること。

```bash
curl -sS -X POST https://context7.com/api/v1/refresh \
  -H "Authorization: Bearer $CONTEXT7_API_KEY" -H 'Content-Type: application/json' \
  -d '{"libraryName":"/ideamans/mautic-cli"}'

# 数分待ってから、llmdocs が入ったか確認する。README.md しか出なければ未解決
curl -sS -G https://context7.com/api/v2/context \
  --data-urlencode 'libraryId=/ideamans/mautic-cli' \
  --data-urlencode 'query=segments are called lists' \
  -H "Authorization: Bearer $CONTEXT7_API_KEY" | grep '^Source:'
```

`CONTEXT7_API_KEY` が無ければ <https://context7.com/add-library> のフォームから。

## Verifying the published result

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install mautic-cli@ideamans-plugins
/mautic-install
/mautic-usage
```

**install スキルの手順を、実際に公開されたリリースに対して一度そのまま実行する
こと。** アセット名（goreleaser の `project_name` は `mautic-cli`、バージョンに
`v` は付かない）、`x86_64`→`amd64` の変換、アーカイブ内のバイナリ名が `mautic`
であること — この3つは公開物に対してしか検証できない。

## gh skill

同じ SKILL.md がそのまま使える。

```bash
gh skill install ideamans/mautic-cli/plugins/mautic-cli/skills/mautic-usage --agent claude-code
gh skill install ideamans/mautic-cli/plugins/mautic-cli/skills/mautic-usage --agent copilot
```
