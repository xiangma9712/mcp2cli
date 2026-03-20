# mcp2cli - Intent

Remote MCP (Model Context Protocol) サーバーが提供するツールを、ローカルで実行可能な CLI コマンドに変換するツール。

## 目的

- MCP の `tools/list` で取得できるツール定義（名前・説明・入力スキーマ）を CLI インターフェースに自動変換し、ローカルから実行可能にする
- `<tool-name> auth login` / `auth logout` で OAuth 等の認証フローを提供する

## 提供形態

### 1. Go パッケージ (`mcp2cli`)

ツール名や MCP サーバーの URL を指定して、5 行程度の Go コードで CLI を構築できるライブラリ。

```go
func main() {
    cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp")
    cli.Run(os.Args)
}
```

カスタマイズ:
- ヘルプテキストの追記・上書き
- 特定ツールの非公開（隠蔽）
- ツール定義の上書き（引数の追加・変更など）

### 2. mcp2cli-runner（汎用実行バイナリ）

`install` コマンドでツール名と URL を指定すると、CLI がインストールされたかのように振る舞う。

```bash
mcp2cli-runner install --name my-tool --url https://mcp.example.com/mcp
```

実態:
- シェルエイリアス (`alias my-tool='mcp2cli-runner run my-tool'`) を登録
- ツールごとの設定は `~/.config/mcp2cli/<tool-name>/` に配置

## 動作イメージ

```
# インストール
$ mcp2cli-runner install --name github-tool --url https://mcp.github.com/mcp

# 認証
$ github-tool auth login

# ツール一覧（MCP tools/list から自動生成）
$ github-tool --help

# ツール実行（MCP tools/call に変換）
$ github-tool create-issue --repo owner/repo --title "Bug report" --body "..."
```

## 技術要素

- MCP プロトコル (Streamable HTTP transport) によるリモートサーバーとの通信
- `tools/list` レスポンスの JSON Schema → CLI フラグへの変換
- OAuth 2.1 認証フローのサポート
- 設定の永続化 (`~/.config/mcp2cli/`)
