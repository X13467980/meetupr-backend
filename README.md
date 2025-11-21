# MeetUP+R Backend

OIC（立命館大学大阪いばらきキャンパス）の学生間、特に留学生と日本人学生の交流を促進するコミュニティアプリのバックエンドAPIサーバーです。

## 📋 概要

MeetUP+Rは、共通の趣味や言語を通じて、自然な出会いや会話のきっかけを作り出すことを目的としたアプリケーションです。

### 主な機能

- ✅ **認証機能**: Auth0によるJWT認証
- ✅ **ユーザー管理**: プロフィール作成・更新・検索
- ✅ **検索機能**: 趣味・興味・言語によるユーザー検索
- ✅ **チャット機能**: WebSocketによるリアルタイムチャット
- ✅ **興味・趣味管理**: マスターデータの取得

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
| GET | `/api/v1/chats` | 参加中のチャット一覧取得 |
| GET | `/api/v1/chats/{chatId}/messages` | チャットメッセージ取得 |

### WebSocket

| エンドポイント | 説明 |
|--------------|------|
| `/ws/chat/{chatID}` | リアルタイムチャット接続 |

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

## 📖 ドキュメント

- [API仕様書](./docs/API_SPECIFICATION.md)
- [データベース設計](./docs/DATABASE.md)
- [要件定義書](./docs/REQUIREMENTS.md)
- [実装状況](./docs/IMPLEMENTATION_STATUS.md)
- [フロントエンドWebSocket実装ガイド](./docs/FRONTEND_WEBSOCKET_IMPLEMENTATION.md)

## 🛠️ 開発

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

詳細は [scripts/README.md](./scripts/README.md) を参照してください。

## 📝 ライセンス

このプロジェクトは内部プロジェクトです。

## 👥 開発チーム

- **フロントエンド**: 秋田／ゆいちゃん／ゆいゆい／みきちゃん
- **バックエンド**: さめ／よーた

## 🔗 関連リンク

- [Swagger UI](http://localhost:8080/swagger/index.html) (サーバー起動時)
- [Supabase](https://supabase.com/)
- [Auth0](https://auth0.com/)
- [Echo Framework](https://echo.labstack.com/)
