# Shigoto-Flow

[![CI](https://github.com/akaitigo/shigoto-flow/actions/workflows/ci.yml/badge.svg)](https://github.com/akaitigo/shigoto-flow/actions/workflows/ci.yml)

**書かない日報、始めよう** — 日報・週報・月報の作成を自動化する中小企業向けレポート自動生成ツール

Googleカレンダー・Slack・GitHub・Gmailから活動データを自動集約し、日報テンプレートに自動流し込み。確認・修正して送信するだけで日報が完成します。週報・月報はAIが日報から自動要約。

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| フロントエンド | Next.js (TypeScript) |
| バックエンド | Go |
| データベース | PostgreSQL |
| 認証 | OAuth2 (Google/Slack/GitHub) |
| AI要約 | Claude API |
| インフラ | Docker Compose / GCP Cloud Run |

## アーキテクチャ

```
┌─────────────────┐     ┌──────────────────┐
│  Next.js (UI)   │────▶│  Go API Server   │
│  - レポート編集  │     │  - OAuth2認証     │
│  - 設定画面     │     │  - データ集約     │
│  - 履歴閲覧     │     │  - レポート生成   │
└─────────────────┘     │  - AI要約        │
                        └────────┬─────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                       │
    ┌─────▼─────┐         ┌─────▼─────┐          ┌─────▼─────┐
    │PostgreSQL │         │Claude API │          │外部サービス│
    └───────────┘         └───────────┘          │Calendar   │
                                                 │Slack      │
                                                 │GitHub     │
                                                 │Gmail      │
                                                 └───────────┘
```

## クイックスタート

### 前提条件

- Go 1.23+
- Node.js 22+
- pnpm 9+
- Docker & Docker Compose
- PostgreSQL 16

### セットアップ

```bash
# リポジトリをクローン
git clone git@github.com:akaitigo/shigoto-flow.git
cd shigoto-flow

# 環境変数を設定
cp .env.example .env
# .env を編集してOAuth2クライアントID等を設定

# データベースを起動
docker compose up -d

# バックエンドを起動
cd backend
go mod download
go run ./cmd/server

# フロントエンドを起動（別ターミナル）
cd frontend
pnpm install
pnpm dev
```

http://localhost:3000 にアクセスしてください。

### コマンド

```bash
make build    # 全体ビルド
make test     # 全テスト実行
make lint     # lint実行
make format   # フォーマット
make check    # lint + test + build
```

## 主な機能

### 自動集約
Google Calendar、Slack、GitHub、Gmailから当日の活動を自動で収集します。OAuth2認証で安全に接続。

### 日報自動生成
集約したデータをテンプレートに自動流し込み。「やったこと」「わかったこと」「次やること」のデフォルトテンプレート付き。カスタマイズも可能。

### 週報・月報自動要約
Claude APIを使って日報から週報を、週報から月報を自動生成。AIが要点を整理し、上長への報告を効率化。

### セキュリティ
- OAuth2トークンはAES-256-GCMで暗号化保存
- 認証ミドルウェアで全APIを保護
- CORS設定で不正アクセスを防止

## デモ

（デモ環境は準備中）

## ライセンス

MIT
