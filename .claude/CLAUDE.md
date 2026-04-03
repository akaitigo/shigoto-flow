# Shigoto-Flow — Agent Instructions

## アーキテクチャ
- モノレポ: frontend/ (Next.js) + backend/ (Go)
- backend は cmd/server/main.go がエントリポイント
- internal/ 以下にドメインロジック: collector, report, summary, handler, model, repository

## Go 規約
- エラーは必ず処理。`_` で握りつぶさない
- context.Context を第一引数に渡す
- テーブル駆動テストを使用
- golangci-lint (.golangci.yml) に従う

## TypeScript 規約
- any 禁止。unknown + 型ガードを使う
- oxlint + biome でlint/format
- vitest でテスト

## API設計
- RESTful JSON API
- エンドポイント: /api/v1/...
- エラーレスポンス: {"error": "message", "code": "ERROR_CODE"}

## 環境変数
- DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME — PostgreSQL接続
- GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET — Google OAuth2
- SLACK_CLIENT_ID, SLACK_CLIENT_SECRET — Slack OAuth2
- GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET — GitHub OAuth2
- ANTHROPIC_API_KEY — Claude API
- PORT — APIサーバーポート (default: 8080)
- FRONTEND_URL — フロントエンドURL (CORS設定用)

## データソース連携
- Google Calendar API v3 — 予定取得
- Slack Web API — メッセージ取得
- GitHub REST API — コミット/PR/Issue取得
- Gmail API — メール送受信サマリー
- 全てOAuth2フローで認証

## テンプレート
- 日報テンプレートは templates テーブルで管理
- デフォルトテンプレート: やったこと/わかったこと/次やること
- ユーザーカスタム可能（セクション追加/削除/並べ替え）
