# MeetUP+R Backend

OIC（立命館大学大阪いばらきキャンパス）の学生間、特に留学生と日本人学生の交流を促進するコミュニティアプリのバックエンドAPIサーバーです。

## 📋 概要

MeetUP+Rは、共通の趣味や言語を通じて、自然な出会いや会話のきっかけを作り出すことを目的としたアプリケーションです。

### 主な機能

- ✅ **認証機能**: Auth0によるJWT認証（Authorization ヘッダーまたはクエリパラメータ対応）
- ✅ **ユーザー管理**: プロフィール作成・更新・検索（アバター画像対応）
- ✅ **検索機能**: キーワード・言語・国による高度なユーザー検索（並列処理で高速化）
- ✅ **チャット機能**: WebSocketによるリアルタイムチャット（メッセージ履歴対応）
- ✅ **興味・趣味管理**: マスターデータの取得

### デプロイ状況

- **本番環境**: Render（`https://meetupr-backend.onrender.com`）
- **フロントエンド**: Vercel（`https://meetupr-frontend.vercel.app`）

## 🛠️ 技術スタック

- **言語**: Go 1.25.3
- **フレームワーク**: Echo v4
- **データベース**: Supabase (PostgreSQL)
- **認証**: Auth0 (JWT)
- **リアルタイム通信**: WebSocket (Gorilla WebSocket)
- **API仕様**: Swagger/OpenAPI

## 📁 プロジェクト構造

```
meetupr-backend/
├── cmd/
│   └── meetupr-backend/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── auth/                    # 認証関連
│   │   └── auth.go
│   ├── db/                      # データベース操作
│   │   └── db.go
│   ├── handlers/                # HTTPハンドラー
│   │   ├── user.go
│   │   ├── chat.go
│   │   ├── interests.go
│   │   └── websocket.go
│   ├── models/                  # データモデル
│   │   ├── user.go
│   │   └── chat.go
│   └── middleware/              # ミドルウェア
├── docs/                        # ドキュメント
│   ├── API_SPECIFICATION.md
│   ├── DATABASE.md
│   ├── REQUIREMENTS.md
│   └── swagger.yaml
├── scripts/                     # ユーティリティスクリプト
│   ├── seed.go
│   └── test_api.sh
└── build/                       # ビルド関連
    └── Dockerfile
```

## 🚀 セットアップ

### 前提条件

- Go 1.25.3 以上
- Supabase アカウントとプロジェクト
- Auth0 アカウントとアプリケーション

### インストール

1. リポジトリをクローン
```bash
git clone <repository-url>
cd meetupr-backend
```

2. 依存関係をインストール
```bash
go mod download
```

3. 環境変数を設定

`.env`ファイルを作成し、以下の環境変数を設定してください：

```env
# Auth0設定
AUTH0_DOMAIN=your-domain.auth0.com
AUTH0_AUDIENCE=your-audience

# Supabase設定
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-supabase-anon-key

# CORS設定（オプション、開発環境では未設定で全許可）
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:3001
```

### 実行

```bash
go run cmd/meetupr-backend/main.go
```

サーバーは `http://localhost:8080` で起動します。

## 📚 APIエンドポイント

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
    "interests": [...]
  }
]
```

### WebSocket

| エンドポイント | 説明 |
|--------------|------|
| `/ws/chat/{chatID}?token={JWT_TOKEN}` | リアルタイムチャット接続（JWTトークンはクエリパラメータで送信） |

詳細なAPI仕様は [API_SPECIFICATION.md](./docs/API_SPECIFICATION.md) または Swagger UI (`http://localhost:8080/swagger/index.html`) を参照してください。

## 🔐 認証

すべてのAPIエンドポイントはAuth0のJWTトークンによる認証が必要です。

リクエストヘッダーに以下の形式でトークンを付与してください：

```
Authorization: Bearer <YOUR_AUTH0_JWT>
```

### 開発モード

開発環境では、環境変数 `DEV_MODE=true` を設定することで、`X-Test-User-ID` ヘッダーを使用したテスト認証が可能です。

```bash
curl -H "X-Test-User-ID: auth0|test_user_id" http://localhost:8080/api/v1/users/me
```

## 🧪 テスト

### APIテスト

```bash
# テストスクリプトの実行
./scripts/test_api.sh <user_id> <email> [interest_id]
```

詳細は [TEST_GUIDE.md](./TEST_GUIDE.md) を参照してください。

### WebSocketテスト

```bash
# WebSocketテストクライアントの実行
go run test_websocket_client.go
```

または、ブラウザで `test_websocket.html` を開いてテストできます。

詳細は [README_WEBSOCKET_TEST.md](./README_WEBSOCKET_TEST.md) を参照してください。

## 🐳 Docker

### ビルド

```bash
docker build -t meetupr-backend -f build/Dockerfile .
```

### 実行

```bash
docker run -p 8080:8080 --env-file .env meetupr-backend
```

## 🚀 デプロイ

### Renderへのデプロイ

詳細な手順は [Renderデプロイ手順](./docs/RENDER_DEPLOYMENT.md) を参照してください。

**クイックスタート**:
1. RenderでWebサービスを作成
2. GitHubリポジトリを接続
3. 環境変数を設定（`AUTH0_DOMAIN`, `AUTH0_AUDIENCE`, `SUPABASE_URL`, `SUPABASE_KEY`, `CORS_ALLOW_ORIGINS`）
4. デプロイ

### Auth0設定

本番環境デプロイ時は、Auth0の設定も更新が必要です。

詳細は [Auth0デプロイ設定ガイド](./docs/AUTH0_DEPLOYMENT_SETUP.md) を参照してください。

**主な設定項目**:
- Application: Allowed Callback URLs, Allowed Logout URLs, Allowed Web Origins
- API: Identifier (Audience) の確認

## 📖 ドキュメント

- [API仕様書](./docs/API_SPECIFICATION.md)
- [データベース設計](./docs/DATABASE.md)
- [要件定義書](./docs/REQUIREMENTS.md)
- [実装状況](./docs/IMPLEMENTATION_STATUS.md)
- [フロントエンドWebSocket実装ガイド](./docs/FRONTEND_WEBSOCKET_GUIDE.md)
- [Renderデプロイ手順](./docs/RENDER_DEPLOYMENT.md)
- [Auth0デプロイ設定ガイド](./docs/AUTH0_DEPLOYMENT_SETUP.md)

## 🛠️ 開発

### サーバー起動

```bash
go run cmd/meetupr-backend/main.go
```

または、ポート番号を環境変数で指定：

```bash
PORT=8080 go run cmd/meetupr-backend/main.go
```

### コード生成

Swaggerドキュメントを更新した場合：

```bash
swag init -g cmd/meetupr-backend/main.go
```

### データベースシード

テストデータを作成する場合：

```bash
go run scripts/seed.go
```

### ユーティリティスクリプト

`scripts/` ディレクトリには以下のユーティリティスクリプトがあります：

- `seed.go`: テストデータの作成
- `create_test_user.go`: テストユーザーの作成
- `create_test_chat_data.go`: テストチャットデータの作成
- `create_test_messages.go`: テストメッセージの作成
- `check_chats.go`: チャットデータの確認
- `check_avatar_url.go`: ユーザーのアバターURL確認
- `test_get_or_create_chat.go`: チャット作成APIのテスト

詳細は [scripts/README.md](./scripts/README.md) を参照してください。

## ⚡ パフォーマンス最適化

### 検索APIの最適化

- **並列処理**: goroutineを使用してユーザー情報の取得を並列化
- **クエリ最適化**: 必要なフィールドのみを個別に取得（Supabaseクライアントの制限対応）
- **興味情報の制限**: フィルター条件がない場合は興味情報を省略

### チャット一覧APIの最適化

- **クエリ最適化**: チャットIDを直接取得してから詳細情報を取得
- **最終メッセージの最適化**: 全メッセージを取得せず、最後のメッセージのみを取得

## 📝 ライセンス

このプロジェクトは内部プロジェクトです。

## 👥 開発チーム

- **フロントエンド**: 秋田／ゆいちゃん／ゆいゆい／みきちゃん
- **バックエンド**: さめ／よーた

## 🔗 関連リンク

### 開発ツール
- [Swagger UI](http://localhost:8080/swagger/index.html) (サーバー起動時)
- [Supabase Dashboard](https://supabase.com/dashboard)
- [Auth0 Dashboard](https://manage.auth0.com/)

### ドキュメント
- [Echo Framework](https://echo.labstack.com/)
- [Supabase Go Client](https://github.com/nedpals/supabase-go)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)

### デプロイ
- [Render](https://render.com/)
- [Vercel](https://vercel.com/)

## 📊 実装状況

### ✅ 実装済み機能

- ✅ ユーザー登録・プロフィール管理
- ✅ ユーザー検索（キーワード・言語・国によるフィルタリング）
- ✅ チャット機能（REST API + WebSocket）
- ✅ 興味・趣味マスタ取得
- ✅ アバター画像対応
- ✅ ネイティブ言語フィールド対応

### ⚠️ 部分的実装

- ⚠️ AIテーマ提案: データベースにフィールドは存在するが、API未実装

### ❌ 未実装機能

- ❌ 匿名「会いたい」ボタン
- ❌ イベント・ミッション提示

詳細は [実装状況](./docs/IMPLEMENTATION_STATUS.md) を参照してください。
