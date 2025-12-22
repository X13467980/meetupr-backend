# MeetUP+R Backend

OIC（立命館大学大阪いばらきキャンパス）の学生間、特に留学生と日本人学生の交流を促進するコミュニティアプリのバックエンドAPIサーバーです。

## プロジェクト概要

**MeetUP+R**は、共通の趣味や言語を通じて、自然な出会いや会話のきっかけを作り出すコミュニティアプリです。

### 主な目的

- 留学生が日本人と関わる機会を増やす
- 言語学習や文化交流を促進する
- 学内コミュニティの国際的つながりを強化する

### 主要機能

- **認証機能**: Auth0を使用したJWT認証（Authorization ヘッダーまたはクエリパラメータ対応）
- **ユーザー管理**: プロフィール作成・更新・検索（アバター画像対応）
- **検索機能**: キーワード・言語・国による高度なユーザー検索（並列処理で高速化、ネイティブ言語フィルタリング対応）
- **チャット機能**: WebSocketによるリアルタイムテキストチャット（メッセージ履歴対応）
- **興味・趣味管理**: マスターデータの取得

詳細な機能要件については、[プロダクト要件定義書](./docs/REQUIREMENTS.md)を参照してください。

## 技術スタック

<div align="center">

![Go](https://img.shields.io/badge/Go-1.25.3-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Echo](https://img.shields.io/badge/Echo-4-00DC82?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Supabase](https://img.shields.io/badge/Supabase-3ECF8E?style=for-the-badge&logo=supabase&logoColor=white)
![Auth0](https://img.shields.io/badge/Auth0-EB5424?style=for-the-badge&logo=auth0&logoColor=white)
![WebSocket](https://img.shields.io/badge/WebSocket-010101?style=for-the-badge&logo=socket.io&logoColor=white)
![Render](https://img.shields.io/badge/Render-46E3B7?style=for-the-badge&logo=render&logoColor=white)

</div>

- **言語**: [Go 1.25.3](https://go.dev/)
- **フレームワーク**: [Echo v4](https://echo.labstack.com/)
- **データベース**: [Supabase](https://supabase.com/) (PostgreSQL)
- **認証**: [Auth0](https://auth0.com/) (JWT)
- **リアルタイム通信**: [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **API仕様**: Swagger/OpenAPI
- **デプロイ**: Render

## セットアップ

### 前提条件

- Go 1.25.3 以上
- Supabase アカウントとプロジェクト
- Auth0 アカウントとアプリケーション

### インストール

1. リポジトリをクローン

```bash
git clone https://github.com/X13467980/meetupr-backend.git
cd meetupr-backend
```

2. 依存関係をインストール

```bash
go mod download
```

3. 環境変数を設定

プロジェクトルートに`.env`ファイルを作成し、以下の環境変数を設定してください：

```env
# Auth0設定（必須）
AUTH0_DOMAIN=your-domain.auth0.com
AUTH0_AUDIENCE=your-audience

# Supabase設定（必須）
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-supabase-anon-key

# CORS設定（オプション、開発環境では未設定で全許可）
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:3001

# ポート設定（オプション、デフォルト: 8080）
PORT=8080

# 認証無効化（開発用、オプション）
DISABLE_AUTH=false
```

**必須の環境変数:**
- `AUTH0_DOMAIN`: Auth0ドメイン
- `AUTH0_AUDIENCE`: Auth0 API Audience
- `SUPABASE_URL`: SupabaseプロジェクトのURL
- `SUPABASE_KEY`: SupabaseのAnon Key

**オプションの環境変数:**
- `CORS_ALLOW_ORIGINS`: CORS許可オリジン（カンマ区切り、未設定時は全許可）
- `PORT`: サーバーポート（デフォルト: 8080）
- `DISABLE_AUTH`: 認証を無効化（開発用、デフォルト: false）

## 開発サーバーの起動

```bash
# 開発サーバーを起動（http://localhost:8080）
go run cmd/meetupr-backend/main.go
```

または、ポート番号を環境変数で指定：

```bash
PORT=8080 go run cmd/meetupr-backend/main.go
```

サーバーは `http://localhost:8080` で起動します。

Swagger UIは `http://localhost:8080/swagger/index.html` で確認できます。

## プロジェクト構造

```
meetupr-backend/
├── cmd/
│   └── meetupr-backend/
│       └── main.go              # エントリーポイント（ルーティング、ミドルウェア設定）
├── internal/
│   ├── auth/                    # 認証関連
│   │   └── auth.go              # Auth0 JWT認証ミドルウェア
│   ├── db/                      # データベース操作
│   │   └── db.go                # Supabaseクライアント、DB操作関数
│   ├── handlers/                # HTTPハンドラー
│   │   ├── user.go              # ユーザー関連APIハンドラー
│   │   ├── chat.go              # チャット関連APIハンドラー
│   │   ├── search.go            # 検索APIハンドラー
│   │   ├── interests.go         # 興味・趣味APIハンドラー
│   │   └── websocket.go         # WebSocketハンドラー
│   ├── models/                  # データモデル
│   │   ├── user.go              # ユーザー・プロフィールモデル
│   │   └── chat.go              # チャット・メッセージモデル
│   └── middleware/              # ミドルウェア
├── docs/                        # ドキュメント
│   ├── API_SPECIFICATION.md     # API仕様書
│   ├── DATABASE.md              # データベース設計書
│   ├── REQUIREMENTS.md          # プロダクト要件定義書
│   ├── FRONTEND_WEBSOCKET_GUIDE.md # フロントエンドWebSocket実装ガイド
│   └── swagger.yaml             # Swagger仕様書
├── scripts/                     # ユーティリティスクリプト
│   ├── seed.go                  # テストデータの作成
│   ├── create_test_user.go      # テストユーザーの作成
│   ├── create_test_chat_data.go # テストチャットデータの作成
│   ├── create_test_messages.go  # テストメッセージの作成
│   ├── check_chats.go           # チャットデータの確認
│   ├── check_avatar_url.go      # ユーザーのアバターURL確認
│   └── test_get_or_create_chat.go # チャット作成APIのテスト
├── build/                       # ビルド関連
│   └── Dockerfile               # Dockerfile
├── render.yaml                  # Renderデプロイ設定
├── go.mod                       # Go依存関係管理
├── go.sum                       # Go依存関係チェックサム
└── README.md                    # このファイル
```

## ドキュメント

- [プロダクト要件定義書](./docs/REQUIREMENTS.md)
- [API仕様書](./docs/API_SPECIFICATION.md)
- [データベース設計書](./docs/DATABASE.md)
- [フロントエンドWebSocket実装ガイド](./docs/FRONTEND_WEBSOCKET_GUIDE.md)

## 主要機能の詳細

### 認証機能

- Auth0を使用したJWT認証
- Authorization ヘッダーまたはクエリパラメータでのトークン送信に対応
- WebSocket接続時はクエリパラメータでトークンを送信
- 開発モードでは認証を無効化可能（`DISABLE_AUTH=true`）

### ユーザー管理

- **ユーザー登録**: 新規ユーザーの登録
- **プロフィール取得**: 自分のプロフィールまたは他ユーザーのプロフィール取得
- **プロフィール更新**: ユーザー名、コメント、言語、興味・趣味、出身地などの更新
- **アバター画像**: Supabase Storageに保存されたアバター画像のURLを取得

### 検索機能

- **キーワード検索**: ユーザー名による検索
- **言語検索**: ネイティブ言語によるフィルタリング（ISO 639-1コード、例: "ja", "en"）
- **国検索**: 出身国によるフィルタリング（ISO 3166-1 alpha-2コード、例: "CN", "US"）
- **統合フィルター**: 言語と国を組み合わせて検索
- **並列処理**: goroutineを使用してユーザー情報の取得を並列化（最大10並列）
- **レスポンス**: `user_id`, `username`, `comment`, `residence`, `avatar_url`, `native_language`, `interests`を含む

### チャット機能

- **チャット一覧取得**: 参加中のチャットルーム一覧（`other_user`に`avatar_url`を含む）
- **チャット作成**: 指定ユーザーとのチャットを取得または作成
- **メッセージ履歴**: 過去のメッセージを取得
- **WebSocket通信**: リアルタイムなメッセージ送受信
- **メッセージ送信**: テキストメッセージの送信と保存
- **セキュリティ**: チャット参加者のみアクセス可能

### 興味・趣味管理

- **マスターデータ取得**: 利用可能な興味・趣味の一覧を取得
- **多言語対応**: 日本語・英語の名称に対応

## APIエンドポイント

### ユーザー (`/api/v1/users`)

| メソッド | エンドポイント | 説明 |
|---------|--------------|------|
| POST | `/api/v1/users/register` | ユーザー登録 |
| GET | `/api/v1/users/me` | 自分のプロフィール取得 |
| PUT | `/api/v1/users/me` | プロフィール更新 |
| GET | `/api/v1/users` | ユーザー検索（クエリパラメータ: `interest_id`, `learning_language`, `spoken_language`） |
| GET | `/api/v1/users/{userId}` | 特定ユーザーのプロフィール取得 |

### 興味・趣味 (`/api/v1/interests`)

| メソッド | エンドポイント | 説明 |
|---------|--------------|------|
| GET | `/api/v1/interests` | 興味・趣味のマスターデータ取得 |

### チャット (`/api/v1/chats`)

| メソッド | エンドポイント | 説明 |
|---------|--------------|------|
| GET | `/api/v1/chats` | 参加中のチャット一覧取得（`other_user`に`avatar_url`を含む） |
| GET | `/api/v1/chats/with/{otherUserId}` | チャットの取得または作成 |
| GET | `/api/v1/chats/{chatId}` | チャット詳細取得 |
| GET | `/api/v1/chats/{chatId}/messages` | チャットメッセージ取得 |

### 検索 (`/api/v1/search/users`)

| メソッド | エンドポイント | 説明 |
|---------|--------------|------|
| GET | `/api/v1/search/users` | ユーザー検索（クエリパラメータ版） |
| POST | `/api/v1/search/users` | ユーザー検索（リクエストボディ版、推奨） |

**リクエスト例（POST）**:
```json
{
  "keyword": "ユーザー名",
  "languages": ["日本語", "英語"],
  "countries": ["CN", "US"]
}
```

**レスポンス例**:
```json
[
  {
    "user_id": "auth0|1234567890",
    "username": "testuser",
    "comment": "こんにちは！",
    "residence": "CN",
    "avatar_url": "https://...",
    "native_language": "ja",
    "interests": [
      {
        "id": 1,
        "name": "プログラミング"
      }
    ]
  }
]
```

### WebSocket

| エンドポイント | 説明 |
|--------------|------|
| `/ws/chat/{chatID}?token={JWT_TOKEN}` | リアルタイムチャット接続（JWTトークンはクエリパラメータで送信） |

詳細なAPI仕様は [API_SPECIFICATION.md](./docs/API_SPECIFICATION.md) または Swagger UI (`http://localhost:8080/swagger/index.html`) を参照してください。

## 認証

すべてのAPIエンドポイントはAuth0のJWTトークンによる認証が必要です。

### REST API

リクエストヘッダーに以下の形式でトークンを付与してください：

```
Authorization: Bearer <YOUR_AUTH0_JWT>
```

### WebSocket

WebSocket接続時は、クエリパラメータでトークンを送信します：

```
ws://localhost:8080/ws/chat/{chatID}?token={YOUR_AUTH0_JWT}
```

### 開発モード

開発環境では、環境変数 `DISABLE_AUTH=true` を設定することで、`X-Test-User-ID` ヘッダーを使用したテスト認証が可能です。

```bash
curl -H "X-Test-User-ID: auth0|test_user_id" http://localhost:8080/api/v1/users/me
```

## ビルド

### 本番用ビルド

```bash
go build -o meetupr-backend ./cmd/meetupr-backend
```

### 実行

```bash
./meetupr-backend
```

または、ポート番号を環境変数で指定：

```bash
PORT=8080 ./meetupr-backend
```

## テスト

### APIテスト

```bash
# テストスクリプトの実行
./scripts/test_api.sh <user_id> <email> [interest_id]
```

### WebSocketテスト

```bash
# WebSocketテストクライアントの実行
go run test_websocket_client.go
```

または、ブラウザで `test_websocket.html` を開いてテストできます。

## Docker

### ビルド

```bash
docker build -t meetupr-backend -f build/Dockerfile .
```

### 実行

```bash
docker run -p 8080:8080 --env-file .env meetupr-backend
```

## 開発ガイドライン

### コミットメッセージ

コミットメッセージは英語で記述し、以下の形式に従ってください：

```
feat: 新機能の追加
fix: バグ修正
docs: ドキュメントの更新
refactor: リファクタリング
style: コードスタイルの変更
test: テストの追加・修正
perf: パフォーマンス改善
chore: ビルドプロセスやツールの変更
```

### ブランチ戦略

- `main`: 本番環境用ブランチ
- `develop-v1`: 開発ブランチ
- `feature/*`: 機能追加用ブランチ
- `fix/*`: バグ修正用ブランチ

### コードスタイル

- Goの標準的なコーディング規約に従う
- エラーハンドリングを適切に実装
- 関数は単一責任の原則に従う
- コメントは英語で記述（公開APIは日本語でも可）

## パフォーマンス最適化

### 検索APIの最適化

- **並列処理**: goroutineを使用してユーザー情報の取得を並列化（最大10並列）
- **クエリ最適化**: 必要なフィールドのみを個別に取得（Supabaseクライアントの制限対応）
- **興味情報の制限**: フィルター条件がない場合は興味情報を省略、フィルターがある場合は最大3件まで取得

### チャット一覧APIの最適化

- **クエリ最適化**: チャットIDを直接取得してから詳細情報を取得（N+1問題の解決）
- **最終メッセージの最適化**: 全メッセージを取得せず、最後のメッセージのみを取得
- **アバター画像の取得**: プロフィール情報と分離して取得（Supabaseクライアントの制限対応）

## トラブルシューティング

### よくある問題

1. **Auth0認証エラー**
   - `.env`ファイルの環境変数を確認
   - Auth0 Dashboardの設定を確認
   - JWTトークンが有効か確認

2. **Supabase接続エラー**
   - `SUPABASE_URL`と`SUPABASE_KEY`が正しく設定されているか確認
   - Supabaseプロジェクトがアクティブか確認
   - ネットワーク接続を確認

3. **WebSocket接続エラー**
   - バックエンドサーバーが起動しているか確認
   - JWTトークンがクエリパラメータで正しく送信されているか確認
   - CORS設定を確認

4. **検索結果が返らない**
   - ブラウザのコンソールでエラーログを確認
   - バックエンドAPIのログを確認
   - ネットワークタブでAPIリクエストを確認
   - Supabaseのデータベースにデータが存在するか確認

5. **`unexpected end of JSON input`エラー**
   - Supabaseクライアントの制限によるエラー（既に対応済み）
   - 個別フィールド取得方式を使用していることを確認
   - エラーハンドリングが適切に実装されているか確認

6. **チャット一覧の読み込みが遅い**
   - データベースクエリの最適化を確認
   - N+1問題が解決されているか確認
   - 並列処理が適切に実装されているか確認

## 開発体制

- **フロントエンド**: 
  - [@kayakku06](https://github.com/kayakku06) - Yuto Akita
  - [@yui949](https://github.com/yui949) - Yui Nishimura
  - [@yui-119](https://github.com/yui-119) - Yui Shimamoto
  - [@mikimiki1207](https://github.com/mikimiki1207) - Miki Fujita
- **バックエンド**: 
  - [@X13467980](https://github.com/X13467980) - Youta Yano
  - [@moyashi0060](https://github.com/moyashi0060) - Ren Sameshima

## ライセンス

このプロジェクトはプライベートプロジェクトです。

## コントリビューション

このプロジェクトはOICの学生向けのプロジェクトです。コントリビューションについては、プロジェクトメンテナーにご連絡ください。

## 開発履歴

- **2025年10月**: プロジェクト開始、認証機能実装
- **2025年11月**: プロフィール管理、検索機能実装、チャット機能実装
- **2025年12月**: パフォーマンス最適化、デプロイ対応、README完成

---

**開発期間**: 2025年10月〜2025年12月

**最終更新**: 2025年12月
