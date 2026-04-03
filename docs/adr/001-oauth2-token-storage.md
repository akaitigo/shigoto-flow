# ADR-001: OAuth2トークン保存方法

## ステータス
承認済み

## コンテキスト
Shigoto-FlowはGoogle/Slack/GitHub/Gmailの4つの外部サービスと連携する。
各サービスへのアクセスにはOAuth2のaccess_tokenとrefresh_tokenが必要。
これらのトークンは機密情報であり、安全に保存する必要がある。

## 検討した選択肢

### A: PostgreSQL + アプリケーション層暗号化
- access_token/refresh_tokenをAES-256-GCMで暗号化してDBに保存
- 暗号化キーは環境変数で管理
- **利点**: シンプル、追加インフラ不要
- **欠点**: 暗号化キーの管理が必要

### B: HashiCorp Vault
- トークンをVaultに保存
- **利点**: エンタープライズレベルのセキュリティ
- **欠点**: インフラコスト大、MVPには過剰

### C: 暗号化なしでDB保存
- **利点**: 最もシンプル
- **欠点**: DBリークでトークン漏洩

## 決定
**選択肢A: PostgreSQL + AES-256-GCM暗号化** を採用する。

## 理由
- MVPフェーズではインフラの複雑さを最小限に抑えたい
- AES-256-GCMは十分なセキュリティを提供する
- 将来的にVaultへの移行も容易（暗号化レイヤーの差し替え）
- 暗号化キーはTOKEN_ENCRYPTION_KEY環境変数で管理

## 実装詳細
- `internal/auth/crypto.go` に暗号化/復号ロジックを配置
- 暗号化キーは32バイト（AES-256）
- 各トークンにランダムなnonceを付与
- DBにはbase64エンコードした暗号文を保存

## 影響
- 環境変数 `TOKEN_ENCRYPTION_KEY` の追加が必要
- `.env.example` に追記
- デプロイ手順にキー生成の手順を追加
