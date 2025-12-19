# Render デプロイ手順

このドキュメントでは、MeetUP+R バックエンドを Render にデプロイする手順を説明します。

## 📋 前提条件

- Render アカウント（[render.com](https://render.com) で無料アカウント作成可能）
- GitHub リポジトリへのアクセス権限
- Supabase プロジェクトの URL と API キー
- Auth0 の設定情報

## 🚀 デプロイ手順

### 1. Render アカウントの作成

1. [render.com](https://render.com) にアクセス
2. 「Get Started for Free」をクリック
3. GitHub アカウントでサインアップ（推奨）

### 2. 新しい Web サービスを作成

1. Render ダッシュボードで「New +」→「Web Service」を選択
2. GitHub リポジトリを接続
   - 「Connect account」をクリックして GitHub を認証
   - リポジトリ `X13467980/meetupr-backend` を選択
   - 「Connect」をクリック

### 3. サービス設定

以下の設定を行います：

#### 基本設定

- **Name**: `meetupr-backend`（任意の名前）
- **Region**: `Oregon (US West)`（最寄りのリージョンを選択）
- **Branch**: `develop-v1`（またはデプロイしたいブランチ）
- **Root Directory**: （空欄のまま）
- **Runtime**: `Go`
- **Build Command**: `go build -o meetupr-backend ./cmd/meetupr-backend`
- **Start Command**: `./meetupr-backend`

#### 環境変数の設定

「Environment」セクションで以下の環境変数を追加：

| キー | 値 | 説明 |
|------|-----|------|
| `PORT` | `8080` | Render が自動設定（変更不要） |
| `AUTH0_DOMAIN` | `your-domain.auth0.com` | Auth0 のドメイン |
| `AUTH0_AUDIENCE` | `your-audience` | Auth0 の Audience |
| `SUPABASE_URL` | `https://your-project.supabase.co` | Supabase プロジェクト URL |
| `SUPABASE_KEY` | `your-supabase-anon-key` | Supabase の匿名キー |
| `CORS_ALLOW_ORIGINS` | `https://meetupr-frontend.vercel.app` | フロントエンドの URL（カンマ区切り、複数ある場合はカンマで区切る） |

**注意**: 
- `PORT` は Render が自動的に設定するため、手動で設定する必要はありません
- 機密情報（`SUPABASE_KEY`、`AUTH0_AUDIENCE` など）は「Secret」として保存することを推奨

### 4. プランの選択

- **Free Plan**: 無料プラン（スリープモードあり、15分間の無操作でスリープ）
- **Starter Plan**: 有料プラン（常時起動、$7/月）

開発・テスト環境では Free Plan で十分です。

### 5. デプロイの開始

1. 「Create Web Service」をクリック
2. Render が自動的にビルドとデプロイを開始します
3. ビルドログを確認して、エラーがないか確認

### 6. デプロイの確認

デプロイが完了すると、以下のような URL が表示されます：

```
https://meetupr-backend.onrender.com
```

この URL にアクセスして、`Hello, World!` が表示されることを確認してください。

## 🔧 設定ファイル（render.yaml）を使用する場合

プロジェクトルートに `render.yaml` ファイルが含まれている場合、Render が自動的に設定を読み込みます。

### render.yaml の内容

```yaml
services:
  - type: web
    name: meetupr-backend
    env: go
    region: oregon
    plan: free
    buildCommand: go build -o meetupr-backend ./cmd/meetupr-backend
    startCommand: ./meetupr-backend
    envVars:
      - key: PORT
        value: 8080
      - key: AUTH0_DOMAIN
        sync: false
      - key: AUTH0_AUDIENCE
        sync: false
      - key: SUPABASE_URL
        sync: false
      - key: SUPABASE_KEY
        sync: false
      - key: CORS_ALLOW_ORIGINS
        sync: false
```

**注意**: `sync: false` の環境変数は、Render ダッシュボードで手動で設定する必要があります。

## 🔄 自動デプロイの設定

デフォルトでは、選択したブランチにプッシュするたびに自動的にデプロイが実行されます。

### 自動デプロイを無効にする場合

1. サービス設定の「Settings」タブを開く
2. 「Auto-Deploy」セクションで「No」を選択

## 📝 環境変数の更新

環境変数を更新する場合：

1. サービス設定の「Environment」タブを開く
2. 環境変数を追加・編集・削除
3. 「Save Changes」をクリック
4. サービスが自動的に再デプロイされます

## 🐛 トラブルシューティング

### ビルドエラー

- **Go のバージョン**: `go.mod` で指定された Go バージョンが Render でサポートされているか確認
- **依存関係**: `go mod download` が正常に実行されているか確認
- **ビルドログ**: Render のビルドログを確認してエラーメッセージを確認

### 起動エラー

- **ポート番号**: `PORT` 環境変数が正しく設定されているか確認
- **環境変数**: 必要な環境変数がすべて設定されているか確認
- **ログ**: Render のログを確認してエラーメッセージを確認

### 接続エラー

- **CORS設定**: `CORS_ALLOW_ORIGINS` にフロントエンドの URL が正しく設定されているか確認
  - 本番環境: `https://meetupr-frontend.vercel.app`
  - 開発環境も許可する場合: `https://meetupr-frontend.vercel.app,http://localhost:3000`
- **Supabase接続**: `SUPABASE_URL` と `SUPABASE_KEY` が正しいか確認
- **Auth0設定**: `AUTH0_DOMAIN` と `AUTH0_AUDIENCE` が正しいか確認
- **フロントエンドのAPIベースURL**: フロントエンドが正しいバックエンドURLを参照しているか確認（例: `https://meetupr-backend.onrender.com`）

## 🔐 セキュリティのベストプラクティス

1. **環境変数の保護**: 機密情報は Render の「Secret」として保存
2. **CORS設定**: 本番環境では、`CORS_ALLOW_ORIGINS` に許可するオリジンのみを設定
   - 本番環境: `https://meetupr-frontend.vercel.app`
   - 開発環境も許可する場合: `https://meetupr-frontend.vercel.app,http://localhost:3000`
3. **HTTPS**: Render は自動的に HTTPS を提供（無料）
4. **ログ**: 本番環境では、機密情報がログに出力されないように注意

## 📝 フロントエンドとの連携

フロントエンドが `https://meetupr-frontend.vercel.app` にデプロイされている場合：

1. **CORS設定**: Render の環境変数 `CORS_ALLOW_ORIGINS` に `https://meetupr-frontend.vercel.app` を設定
2. **フロントエンドのAPIベースURL**: フロントエンドの環境変数で、バックエンドのURL（例: `https://meetupr-backend.onrender.com`）を設定
3. **WebSocket接続**: フロントエンドのWebSocket接続URLもバックエンドのURLに合わせて設定（例: `wss://meetupr-backend.onrender.com/ws/chat/{chatID}`）

## 🔐 Auth0 設定

フロントエンドとバックエンドを本番環境にデプロイする際、Auth0 の設定も更新する必要があります。

詳細は [Auth0 デプロイ設定ガイド](./AUTH0_DEPLOYMENT_SETUP.md) を参照してください。

### 主な設定項目

1. **Application（フロントエンド用）**:
   - Allowed Callback URLs: `https://meetupr-frontend.vercel.app/callback`
   - Allowed Logout URLs: `https://meetupr-frontend.vercel.app`
   - Allowed Web Origins: `https://meetupr-frontend.vercel.app`

2. **API（バックエンド用）**:
   - Identifier (Audience) が `AUTH0_AUDIENCE` 環境変数と一致していることを確認

## 📚 参考リンク

- [Render ドキュメント](https://render.com/docs)
- [Render Go ガイド](https://render.com/docs/deploy-go)
- [環境変数の管理](https://render.com/docs/environment-variables)
- [Auth0 デプロイ設定ガイド](./AUTH0_DEPLOYMENT_SETUP.md)

## 🆘 サポート

問題が発生した場合：

1. Render のビルドログとログを確認
2. ローカル環境で動作確認
3. GitHub Issues で報告
