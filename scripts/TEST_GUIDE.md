# テストガイド

## 前提条件

1. **環境変数の設定**
   - `.env`ファイルに`SUPABASE_URL`と`SUPABASE_KEY`が設定されていること
   - テスト用に`DISABLE_AUTH=true`を設定することも可能

2. **データベースの準備**
   ```bash
   # テストデータを投入
   go run scripts/seed.go
   ```

## サーバーの起動

### 開発モード（認証なし）

```bash
DISABLE_AUTH=true go run cmd/meetupr-backend/main.go
```

### 本番モード（認証あり）

```bash
go run cmd/meetupr-backend/main.go
```

## APIテスト

### 1. チャット一覧を取得

```bash
# 開発モードの場合
curl -X GET http://localhost:8080/api/v1/chats \
  -H "X-Test-User-ID: auth0|6917784d99703fe24aebd01d" \
  -H "X-Test-User-Email: testuser1@example.com"

# 本番モードの場合（Auth0トークンが必要）
curl -X GET http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer <YOUR_AUTH0_TOKEN>"
```

### 2. チャットメッセージを取得

```bash
# 開発モードの場合
curl -X GET "http://localhost:8080/api/v1/chats/3/messages" \
  -H "X-Test-User-ID: auth0|6917784d99703fe24aebd01d" \
  -H "X-Test-User-Email: testuser1@example.com"

# 本番モードの場合
curl -X GET "http://localhost:8080/api/v1/chats/3/messages" \
  -H "Authorization: Bearer <YOUR_AUTH0_TOKEN>"
```

### 3. テストスクリプトを使用

```bash
# チャット一覧とメッセージを取得
./scripts/test_api.sh "auth0|6917784d99703fe24aebd01d" "testuser1@example.com" 3
```

## データベースの確認

```bash
# チャットデータを確認
go run scripts/check_chats.go
```

## トラブルシューティング

### チャット一覧が空の場合

1. **データベースにデータが存在するか確認**
   ```bash
   go run scripts/check_chats.go
   ```

2. **seedスクリプトを再実行**
   ```bash
   go run scripts/seed.go
   ```

3. **サーバーログを確認**
   - `GetUserChats: fetching chats for user ...`
   - `GetUserChats: results1=..., results2=...`
   - エラーメッセージがないか確認

### 認証エラーの場合

- サーバーが`DISABLE_AUTH=true`モードで起動しているか確認
- 本番モードの場合は、有効なAuth0トークンが必要

### "You are not a participant in this chat"エラーの場合

- 指定したユーザーIDがチャットの参加者であることを確認
- チャットIDが正しいか確認

