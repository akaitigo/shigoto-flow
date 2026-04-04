# Harvest Report: Shigoto-Flow (#3227)

## プロジェクト概要
- **名前**: Shigoto-Flow
- **ドメイン**: ノーコード業務自動化
- **概要**: 日報・週報・月報の作成を自動化する中小企業向けレポート自動生成ツール
- **リポジトリ**: https://github.com/akaitigo/shigoto-flow
- **タグ**: v1.0.0
- **パイプライン実行日**: 2026-04-04

## 統計

### 規模
| 指標 | 数値 |
|------|------|
| Go ソースファイル | 47 |
| TypeScript/TSX | 13 |
| テスト (Go) | 45 (7パッケージ) |
| テスト (Frontend) | 8 (2ファイル) |
| テスト合計 | 53 |
| コミット数 | 33 |

### Issues: 5/5 完了 (100%)
| # | タイトル | ラベル |
|---|---------|--------|
| 1 | OAuth2認証フロー実装 | model:opus |
| 2 | 外部データソースコレクター完成 | model:sonnet |
| 3 | 日報自動生成エンジン | model:sonnet |
| 4 | 週報・月報自動要約生成 | model:sonnet |
| 5 | 認証・ユーザー管理とフロントエンド統合 | model:sonnet |

### Pull Requests: 10/10 マージ (100%)
| # | タイトル | 種別 |
|---|---------|------|
| 6 | OAuth2認証フロー + トークン暗号化 | feat |
| 7 | 外部データソースコレクター完成 | feat |
| 8 | 日報自動生成エンジン + 送信サービス | feat |
| 9 | 週報・月報自動要約生成 | feat |
| 10 | 認証ミドルウェア + フロントエンド統合 | feat |
| 11 | v1.0.0 Ship | chore |
| 12 | Review Round 1 | fix(security) |
| 13 | Review Round 2 | fix |
| 14 | Review Round 3 | fix(security) |
| 15 | Review Round 4 | fix |

---

## Review Loop 結果（4ラウンド完走）

### 全体サマリー
| ラウンド | 焦点 | CRITICAL | HIGH | PR |
|---------|------|----------|------|-----|
| R1 | 表面的セキュリティ | 4→0 | 1→0 | #12 |
| R2 | 認証基盤・DI・バリデーション | 4→0 | 2→0 | #13 |
| R3 | JWT認証・プロンプト注入 | 1→0 | 2→0 | #14 |
| R4 | 認証フロー・鍵分離・ヘッダー | 3→0 | 3→0 | #15 |
| **累計** | | **12件修正** | **8件修正** | **4 PR** |

### ラウンド1: 表面的セキュリティ
| # | 重要度 | ファイル | 指摘内容 |
|---|--------|---------|----------|
| 1 | CRITICAL | handler.go | 認証ミドルウェア未接続 |
| 2 | CRITICAL | report.go | IDOR脆弱性（オーナーシップ未チェック） |
| 3 | CRITICAL | config.go | デフォルト暗号化キーがハードコード |
| 4 | CRITICAL | main.go | nilエンクリプタでパニック |
| 5 | HIGH | handler.go | リクエストボディサイズ制限なし |

### ラウンド2: 認証基盤・DI
| # | 重要度 | ファイル | 指摘内容 |
|---|--------|---------|----------|
| 1 | CRITICAL | middleware/auth.go | X-User-IDヘッダーを無検証で信頼 |
| 2 | CRITICAL | main.go | Collector未接続（常に503） |
| 3 | CRITICAL | main.go | backendURLがlocalhost固定 |
| 4 | CRITICAL | report.go | ReportType/Status未検証 |
| 5 | HIGH | collector/slack.go | ok=falseを無視、TS握り潰し |
| 6 | HIGH | config.go | DBパスワードデフォルト値 |

### ラウンド3: JWT認証・プロンプト注入
| # | 重要度 | ファイル | 指摘内容 |
|---|--------|---------|----------|
| 1 | CRITICAL | middleware/auth.go | X-User-ID偽造でバイパス可能 |
| 2 | HIGH | summarizer.go | プロンプトインジェクション |
| 3 | HIGH | datasource.go | provider未検証 |

### ラウンド4: 認証フロー・鍵分離
| # | 重要度 | ファイル | 指摘内容 |
|---|--------|---------|----------|
| 1 | CRITICAL | auth.go | OAuthパス非公開→認証開始不能 |
| 2 | CRITICAL | config.go | JWT署名鍵とAES暗号化鍵が同一 |
| 3 | CRITICAL | generate.go | GenerateReportのType未検証 |
| 4 | HIGH | handler.go | セキュリティヘッダー不在 |
| 5 | HIGH | email.go | SMTPヘッダーインジェクション |
| 6 | HIGH | auth.go | isPublicPathのHasPrefix過一致 |

### 残存MEDIUM/LOW（未修正・Issue化推奨）
| # | 重要度 | 指摘内容 | リスク |
|---|--------|----------|--------|
| 1 | MEDIUM | OAuth stateがインメモリ（再起動で消失） | マルチインスタンスで不整合 |
| 2 | MEDIUM | AES-GCM AAD未使用（ciphertext transplant可能） | DBレコード操作で悪用可能 |
| 3 | MEDIUM | JWT jti/iatクレーム欠如（リプレイ攻撃） | HTTPS緩和だがログ漏洩時に問題 |
| 4 | MEDIUM | 外部APIエラーボディをログに出力 | client_secret漏洩の可能性 |
| 5 | MEDIUM | DB sslmodeデフォルトdisable | 本番で平文通信のリスク |
| 6 | LOW | Template名長さ制限なし | MaxBytesReaderで1MB緩和 |
| 7 | LOW | handler/パッケージのテストなし | カバレッジ不足 |
| 8 | LOW | フロントエンドAPI統合がスタブ | UIから機能到達不能 |

---

## 認証セキュリティの進化過程

本プロジェクトの最大の学びは、認証が4ラウンドかけて段階的に成熟した過程。

```
R0(初期): ミドルウェア自体がルーターに未接続 → 全APIが公開状態
    ↓ R1修正
R1: ミドルウェア接続 → だがX-User-IDヘッダーを無検証で信頼
    ↓ R2修正
R2: Context経由に統一 → だがContext値もヘッダーそのまま（偽造可能）
    ↓ R3修正
R3: HMAC署名トークン検証導入 → だがOAuth自体が認証必須で開始不能
    ↓ R4修正
R4: OAuthを公開パス化 + callbackでJWT発行 + 鍵分離 → 動作するフロー完成
```

**教訓**: 認証は「繋がっている」だけでは不十分。「検証している」「フローが成立する」「鍵が分離されている」の全層を初期設計で担保すべき。

---

## テンプレート改善提案

### A. スキャフォールド時に解決すべき問題（R1-R4で繰り返し指摘）

| # | 問題 | 発生ラウンド | 改善策 |
|---|------|------------|--------|
| 1 | 認証ミドルウェア未接続 | R1 | ルーター生成時に自動接続 |
| 2 | X-User-IDヘッダー直信頼 | R2-R3 | JWT検証をデフォルト実装に含める |
| 3 | サービスDI未接続 | R2 | main.goテンプレートに全サービス初期化を含める |
| 4 | IDOR（オーナーシップ未チェック） | R1 | repository層のGet/Updateに必ずuserIDフィルタを含める |
| 5 | 鍵分離 | R4 | TOKEN_ENCRYPTION_KEY + JWT_SECRET を別環境変数でテンプレート化 |
| 6 | OAuthフローのchicken-and-egg | R4 | OAuthをログイン手段として設計（callbackでJWT発行） |
| 7 | セキュリティヘッダー | R4 | CORSミドルウェアにデフォルト含める |
| 8 | バリデーション | R2-R4 | 列挙型のバリデーション関数をmodelに同梱 |

### B. CI/ツール設定の改善

| # | 問題 | 改善策 |
|---|------|--------|
| 1 | golangci-lint Go版不一致 | `go vet` + 個別linterを直接実行するCIに統一 |
| 2 | pnpm-workspace.yaml | `packages: []` をテンプレートに含める |
| 3 | golangci-lint v1→v2 | `.golangci.yml` に `version: "2"` + formatters分離 |
| 4 | Go module version | go.modの`go`ディレクティブをCI環境と合わせる |

### C. コード品質の改善

| # | 問題 | 改善策 |
|---|------|--------|
| 1 | Claude APIプロンプト結合 | system/userメッセージ分離をデフォルト化 |
| 2 | Slack API ok=false無視 | 外部API応答の成否チェックをコレクタテンプレートに含める |
| 3 | SMTPインジェクション | sender系のCRLF検証をテンプレートに含める |
| 4 | handlerテスト不在 | httptest使用のハンドラテストテンプレートを用意 |

---

## プロセス振り返り

### 効果的だったこと
- **4ラウンドレビュー**: 各ラウンドで視点が深まり、R1では見えなかったchicken-and-egg問題がR4で発見
- **3観点並列レビュー**: コード品質・セキュリティ・破壊的テストの並列実行で網羅性が高い
- **修正→再レビューのループ**: 修正自体が新しい問題を生む（R3のJWT導入→R4のOAuthデッドロック）ことを検出

### 改善すべきこと
- **R1の判定が甘すぎた**: 「CRITICAL 0 / HIGH 0 → ループ終了」と判断したが、表面的な修正のみで深層問題が残存
- **フロントエンド統合が後回し**: バックエンドAPIは完成したがフロントエンドのonClickハンドラ等がスタブのまま
- **E2E検証の欠如**: 個別のユニットテストはパスしているが、ユーザーフローの端から端まで動くかの検証がない
- **レビューの独立性**: 同じコンテキストでコードを書いた直後にレビューするため、バイアスがかかりやすい

### 数値
- **パイプライン全体**: Launch→Build→Ship→Review(4R)→Harvest
- **累計修正**: CRITICAL 12件 + HIGH 8件 = 20件
- **PR数**: Build 5 + Ship 1 + Review 4 = 10 PR
- **テスト**: 53テスト全パス
