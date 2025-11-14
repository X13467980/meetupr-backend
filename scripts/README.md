# テストデータ作成スクリプト

チャット機能のテスト用データを作成するためのスクリプトです。

## 🌱 一括データ投入（Seed）

Railsのseedのように、全てのテストデータを一括で作成します。

```bash
go run scripts/seed.go
```

**作成されるデータ:**
- 4人のテストユーザー
- 4つのチャットルーム
- 複数のメッセージ

**注意:**
- 既に存在するデータはスキップされます（エラーになりません）
- 複数回実行しても安全です（冪等性）

## テストユーザーの作成

テスト用のユーザーを直接データベースに作成します。

```bash
go run scripts/create_test_user.go -id <USER_ID> -email <EMAIL> -username <USERNAME>
```

**例:**
```bash
go run scripts/create_test_user.go \
  -id "auth0|test_user_12345" \
  -email "testuser2@example.com" \
  -username "testuser2"
```

**注意:**
- ユーザーIDはAuth0形式（`auth0|xxxxx`）を推奨しますが、任意の形式でも作成可能です
- EmailとUsernameは一意である必要があります
- 既に存在するEmail/Username/IDの場合はエラーになります

## チャットルームの作成

2人のユーザー間のチャットルームを作成します。

```bash
go run scripts/create_test_chat_data.go -user1 <USER1_ID> -user2 <USER2_ID> [-theme <THEME>]
```

**例:**
```bash
go run scripts/create_test_chat_data.go \
  -user1 "auth0|6917784d99703fe24aebd01d" \
  -user2 "auth0|another_user_id" \
  -theme "ゲームについて話そう"
```

**出力:**
- 作成されたチャットルームのIDが表示されます
- WebSocketテスト用のコマンドも表示されます

## メッセージの作成

既存のチャットルームにメッセージを追加します。

```bash
go run scripts/create_test_messages.go -chat <CHAT_ID> -sender <SENDER_ID> -content <MESSAGE> [-count <NUMBER>]
```

**例:**
```bash
# 1つのメッセージを作成
go run scripts/create_test_messages.go \
  -chat 1 \
  -sender "auth0|6917784d99703fe24aebd01d" \
  -content "こんにちは！"

# 複数のメッセージを作成
go run scripts/create_test_messages.go \
  -chat 1 \
  -sender "auth0|6917784d99703fe24aebd01d" \
  -content "テストメッセージ" \
  -count 5
```

## 完全なテストデータ作成の流れ

1. **2人のテストユーザーを作成**（既に登録済みの場合はスキップ）
   ```bash
   # 1人目のユーザー（既に存在する場合はスキップ）
   # 2人目のテストユーザーを作成
   go run scripts/create_test_user.go \
     -id "auth0|test_user_12345" \
     -email "testuser2@example.com" \
     -username "testuser2"
   ```

2. **チャットルームを作成**
   ```bash
   go run scripts/create_test_chat_data.go \
     -user1 "auth0|user1_id" \
     -user2 "auth0|user2_id"
   ```

3. **メッセージを作成**（オプション）
   ```bash
   go run scripts/create_test_messages.go \
     -chat 1 \
     -sender "auth0|user1_id" \
     -content "初めまして！"
   ```

4. **WebSocketでテスト**
   ```bash
   # ターミナル1
   ./test_ws_client -addr localhost:8080 -chat 1 -token "<USER1_TOKEN>"
   
   # ターミナル2
   ./test_ws_client -addr localhost:8080 -chat 1 -token "<USER2_TOKEN>"
   ```

## 認証なしでAPIをテストする方法

Auth0トークンが不要な開発モードを使用できます。

### 1. サーバーを開発モードで起動

```bash
DISABLE_AUTH=true go run cmd/meetupr-backend/main.go
```

または、`.env`ファイルに追加：
```
DISABLE_AUTH=true
```

### 2. APIリクエストにヘッダーを追加

```bash
# チャット一覧を取得
curl -X GET http://localhost:8080/api/v1/chats \
  -H "X-Test-User-ID: auth0|6917784d99703fe24aebd01d" \
  -H "X-Test-User-Email: testuser1@example.com"

# 別のユーザーでテスト
curl -X GET http://localhost:8080/api/v1/chats \
  -H "X-Test-User-ID: auth0|test_user_67890" \
  -H "X-Test-User-Email: testuser2@example.com"
```

### 3. テストスクリプトを使用

```bash
# デフォルトユーザーでテスト
./scripts/test_api.sh

# 指定したユーザーでテスト
./scripts/test_api.sh "auth0|test_user_67890" "testuser2@example.com"

# チャットメッセージも取得
./scripts/test_api.sh "auth0|6917784d99703fe24aebd01d" "testuser1@example.com" 3
```

**注意**: `DISABLE_AUTH=true`モードでは、`X-Test-User-ID`ヘッダーでユーザーIDを指定できます。ヘッダーがない場合は、デフォルトのテストユーザーIDが使用されます。

## 注意事項

- ユーザーIDは既に`users`テーブルに存在している必要があります
- チャットルームは既に存在する場合、エラーになる可能性があります（一意制約）
- メッセージを作成する前に、チャットルームが存在することを確認してください
- **`DISABLE_AUTH=true`は開発環境でのみ使用してください。本番環境では絶対に使用しないでください。**

