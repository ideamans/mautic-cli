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

```bash
git checkout main && git pull
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

## Registering in the marketplace (first release only)

`ideamans/claude-public-plugins` の `.claude-plugin/marketplace.json` に追加:

```json
{
  "name": "mautic-cli",
  "source": {
    "source": "git-subdir",
    "url": "https://github.com/ideamans/mautic-cli.git",
    "path": "plugins/mautic-cli"
  }
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

`CONTEXT7_API_KEY` が無ければ <https://context7.com/add-library> のフォームから。

## Verifying the published result

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install mautic-cli@ideamans
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
