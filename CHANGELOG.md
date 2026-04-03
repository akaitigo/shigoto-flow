# Changelog

## [1.0.0] - 2026-04-04

### Added
- OAuth2認証フロー（Google/Slack/GitHub/Gmail）
- AES-256-GCMによるトークン暗号化保存
- 外部データソースコレクター（Calendar/Slack/GitHub/Gmail）
- 並行データ収集（goroutine + errgroup）
- 部分的失敗ハンドリング（1ソース失敗でも他は継続）
- 日報自動生成エンジン（テンプレート + データ集約）
- 日報テンプレートのカスタマイズ機能
- 送信サービス（Slack Webhook/SMTP Email）
- 週報・月報のAI自動要約（Claude API連携）
- 要約設定（重点項目/除外項目/長さ/詳細度）
- 認証ミドルウェア（パブリックパス自動バイパス）
- Next.jsフロントエンド（レポート編集/設定ダッシュボード）
- PostgreSQLマイグレーション
- GitHub Actions CI（Go + Next.js）
- ADR-001: OAuth2トークン保存方法の決定
- Docker Compose開発環境
