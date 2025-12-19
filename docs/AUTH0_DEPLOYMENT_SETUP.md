# Auth0 デプロイ設定ガイド

このドキュメントでは、MeetUP+R アプリケーションを本番環境にデプロイする際の Auth0 設定手順を説明します。

## 📋 概要

フロントエンドとバックエンドを本番環境にデプロイする際、Auth0 の設定を更新する必要があります。

- **フロントエンド**: `https://meetupr-frontend.vercel.app`
- **バックエンド**: `https://meetupr-backend.onrender.com` (Render にデプロイ後)

---

## 🔧 Auth0 ダッシュボードでの設定

### 1. Application（フロントエンド用）の設定

Auth0 ダッシュボードで、フロントエンド用の Application を開きます。

#### Settings タブで以下を更新：

##### Allowed Callback URLs
```
https://meetupr-frontend.vercel.app/callback,http://localhost:3000/callback
```
- 本番環境のコールバックURLを追加
- 開発環境も含める場合は `http://localhost:3000/callback` も追加

##### Allowed Logout URLs
```
https://meetupr-frontend.vercel.app,http://localhost:3000
```
- 本番環境のログアウト後のリダイレクト先を追加
- 開発環境も含める場合は `http://localhost:3000` も追加

##### Allowed Web Origins
```
https://meetupr-frontend.vercel.app,http://localhost:3000
```
- 本番環境のオリジンを追加
- 開発環境も含める場合は `http://localhost:3000` も追加

##### Allowed Origins (CORS)
```
https://meetupr-frontend.vercel.app,http://localhost:3000
```
- 本番環境のオリジンを追加
- 開発環境も含める場合は `http://localhost:3000` も追加

**注意**: 各設定はカンマ区切りで複数のURLを指定できます。

---

### 2. API（バックエンド用）の設定

Auth0 ダッシュボードで、バックエンド用の API を開きます。

#### Settings タブで確認：

##### Identifier (Audience)
- この値が `AUTH0_AUDIENCE` 環境変数と一致していることを確認
- 例: `https://meetupr-api.com` または `https://api.meetupr.com`

**重要**: この値は変更しないでください。変更すると既存のトークンが無効になります。

#### その他の設定

- **Token Endpoint Authentication Method**: `None` または `Post`（通常は `None` で問題ありません）
- **Allow Skipping User Consent**: 本番環境では `false` に設定（ユーザー同意を要求）

---

### 3. Rules / Actions の確認（オプション）

カスタムクレーム（例: `https://meetupr.com/email`）を使用している場合、Rules または Actions が正しく設定されているか確認してください。

#### 確認ポイント

- カスタムクレームが正しく追加されているか
- 本番環境でも動作するか（テスト環境と本番環境で異なる設定が必要な場合）

---

## 🔐 環境変数の設定

### フロントエンド（Vercel）

Vercel の環境変数で以下を設定：

| キー | 値 | 説明 |
|------|-----|------|
| `AUTH0_DOMAIN` | `your-domain.auth0.com` | Auth0 のドメイン |
| `AUTH0_CLIENT_ID` | `your-client-id` | Application の Client ID |
| `AUTH0_AUDIENCE` | `your-audience` | API の Identifier (Audience) |
| `AUTH0_CALLBACK_URL` | `https://meetupr-frontend.vercel.app/callback` | コールバックURL |
| `API_BASE_URL` | `https://meetupr-backend.onrender.com` | バックエンドAPIのベースURL |

### バックエンド（Render）

Render の環境変数で以下を設定：

| キー | 値 | 説明 |
|------|-----|------|
| `AUTH0_DOMAIN` | `your-domain.auth0.com` | Auth0 のドメイン |
| `AUTH0_AUDIENCE` | `your-audience` | API の Identifier (Audience) |

---

## ✅ 設定チェックリスト

### Application（フロントエンド用）

- [ ] Allowed Callback URLs に `https://meetupr-frontend.vercel.app/callback` を追加
- [ ] Allowed Logout URLs に `https://meetupr-frontend.vercel.app` を追加
- [ ] Allowed Web Origins に `https://meetupr-frontend.vercel.app` を追加
- [ ] Allowed Origins (CORS) に `https://meetupr-frontend.vercel.app` を追加

### API（バックエンド用）

- [ ] Identifier (Audience) が `AUTH0_AUDIENCE` 環境変数と一致していることを確認
- [ ] Token Endpoint Authentication Method が適切に設定されていることを確認

### 環境変数

- [ ] フロントエンド（Vercel）の環境変数が正しく設定されている
- [ ] バックエンド（Render）の環境変数が正しく設定されている

---

## 🧪 動作確認

### 1. ログイン機能の確認

1. フロントエンド（`https://meetupr-frontend.vercel.app`）にアクセス
2. 「ログイン」をクリック
3. Auth0 のログイン画面が表示されることを確認
4. ログイン後、フロントエンドにリダイレクトされることを確認

### 2. API 呼び出しの確認

1. ブラウザの開発者ツールで Network タブを開く
2. ログイン後、API リクエストが正常に送信されているか確認
3. 401 Unauthorized エラーが発生していないか確認

### 3. WebSocket 接続の確認

1. チャット機能を使用
2. WebSocket 接続が正常に確立されることを確認
3. メッセージの送受信が正常に動作することを確認

---

## 🐛 トラブルシューティング

### ログイン後にリダイレクトされない

**原因**: Allowed Callback URLs に正しいURLが設定されていない

**解決方法**:
1. Auth0 ダッシュボードで Application の Settings を開く
2. Allowed Callback URLs に `https://meetupr-frontend.vercel.app/callback` が含まれているか確認
3. 含まれていない場合は追加して保存

### 401 Unauthorized エラーが発生する

**原因**: 
- `AUTH0_AUDIENCE` が正しく設定されていない
- Token の Audience が API の Identifier と一致していない

**解決方法**:
1. フロントエンドの `AUTH0_AUDIENCE` 環境変数が API の Identifier と一致しているか確認
2. バックエンドの `AUTH0_AUDIENCE` 環境変数が API の Identifier と一致しているか確認
3. Auth0 ダッシュボードで API の Identifier を確認

### CORS エラーが発生する

**原因**: 
- Allowed Web Origins にフロントエンドのURLが設定されていない
- バックエンドの CORS 設定が正しくない

**解決方法**:
1. Auth0 ダッシュボードで Application の Settings を開く
2. Allowed Web Origins に `https://meetupr-frontend.vercel.app` が含まれているか確認
3. バックエンドの `CORS_ALLOW_ORIGINS` 環境変数に `https://meetupr-frontend.vercel.app` が設定されているか確認

---

## 📚 参考リンク

- [Auth0 ドキュメント](https://auth0.com/docs)
- [Auth0 Application Settings](https://auth0.com/docs/get-started/applications)
- [Auth0 API Settings](https://auth0.com/docs/get-started/apis)
- [Auth0 CORS 設定](https://auth0.com/docs/cross-origin-authentication)

---

## 🔄 開発環境と本番環境の切り替え

### 開発環境

- **フロントエンド**: `http://localhost:3000`
- **バックエンド**: `http://localhost:8080`
- **Auth0設定**: 開発用の Application と API を使用（または本番環境の設定に開発環境のURLも追加）

### 本番環境

- **フロントエンド**: `https://meetupr-frontend.vercel.app`
- **バックエンド**: `https://meetupr-backend.onrender.com`
- **Auth0設定**: 本番用の Application と API を使用

**推奨**: 開発環境と本番環境で別々の Auth0 Application と API を使用することを推奨します。

---

## ⚠️ 重要な注意事項

1. **Identifier (Audience) の変更**: API の Identifier を変更すると、既存のトークンが無効になります。変更する場合は、すべてのユーザーが再ログインする必要があります。

2. **Callback URLs のセキュリティ**: 許可する Callback URLs は必要最小限に留めてください。不正なURLを追加すると、セキュリティリスクが発生します。

3. **環境変数の保護**: 機密情報（Client ID、Client Secret など）は環境変数として管理し、コードに直接記述しないでください。

4. **HTTPS の使用**: 本番環境では必ず HTTPS を使用してください。Auth0 も HTTPS を推奨しています。
