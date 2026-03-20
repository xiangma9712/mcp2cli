---
status: fulfilled
date: 2026-03-20
---

# 02: Quality - プロジェクト品質基盤の整備

開発・CI/CD・コード品質の基盤を整え、継続的な開発に耐えうる状態にする。

## パッケージ構成の見直し

- `internal/` を導入し、外部に公開すべきでないパッケージを移動
  - `mcp/` → `internal/mcp/` (MCP クライアント実装)
  - `schema/` → `internal/schema/` (スキーマ変換)
  - `auth/` → `internal/auth/` (認証フロー)
  - `cfgstore/` → `internal/cfgstore/` (設定管理)
- ルートパッケージ (`mcp2cli.go`) のみを公開 API として維持

## ビルド基盤

- `Makefile` の導入
  - `make build` — バイナリビルド
  - `make test` — テスト実行
  - `make lint` — リンター実行
  - `make fmt` — フォーマット
  - `make ci` — CI 向け一括チェック

## CI/CD

- `.github/workflows/ci.yml` の作成
  - Go のビルド・テスト
  - golangci-lint の実行
  - 複数 Go バージョンでのマトリクステスト

## リンター

- `golangci-lint` の導入
- `.golangci.yml` 設定ファイルの作成

## 開発環境

- `mise` による開発ツールのセットアップ (`mise.toml`)
  - Go バージョン
  - golangci-lint
  - その他開発ツール

## Claude Code 設定

- `.claude/settings.json` のコミット
  - プロジェクト固有の許可設定
  - CLAUDE.md によるプロジェクトコンテキスト

## README

- プロジェクト概要
- インストール方法
- Go パッケージとしての使用例
- mcp2cli-runner としての使用例
