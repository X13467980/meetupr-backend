# WebSocketチャット機能のテスト方法

## 前提条件

1. サーバーが起動していること（`go run cmd/meetupr-backend/main.go`）
2. Auth0のJWTトークンを取得していること
3. テスト用のチャットルーム（chat_id）が存在すること

## クイックスタート

### 1. サーバーを起動

```bash
# 環境変数を設定（.envファイルまたは環境変数）
export AUTH0_DOMAIN="your-domain.auth0.com"
export AUTH0_AUDIENCE="your-audience"
export SUPABASE_URL="your-supabase-url"
export SUPABASE_KEY="your-supabase-key"

# サーバーを起動
go run cmd/meetupr-backend/main.go
```

### 2. テストクライアントをビルド

```bash
go build -o test_ws_client test_websocket_client.go
```

### 3. テストを実行

```bash
# 方法A: スクリプトを使用（推奨）
./test_websocket.sh "YOUR_JWT_TOKEN" 1 localhost:8080

# 方法B: 直接実行
./test_ws_client -addr localhost:8080 -chat 1 -token "YOUR_JWT_TOKEN"
```

## テスト方法

### 方法1: Goテストクライアントを使用（推奨）

1. **テストクライアントをビルド**
   ```bash
   go build -o test_ws_client test_websocket_client.go
   ```

2. **JWTトークンを取得**
   - Auth0のダッシュボードからテスト用トークンを取得
   - または、フロントエンドアプリから取得したトークンを使用
   - Postmanやcurlで `/api/v1/users/register` を呼び出してトークンを取得

3. **テストクライアントを実行**
   ```bash
   ./test_ws_client -addr localhost:8080 -chat 1 -token "YOUR_JWT_TOKEN"
   ```

   パラメータ:
   - `-addr`: サーバーアドレス（デフォルト: localhost:8080）
   - `-chat`: チャットID（デフォルト: 1）
   - `-token`: JWTトークン（必須）

4. **複数のクライアントでテスト**
   別のターミナルで別のユーザートークンを使用して接続し、メッセージの送受信を確認:
   ```bash
   # ターミナル1
   ./test_ws_client -addr localhost:8080 -chat 1 -token "USER1_JWT_TOKEN"
   
   # ターミナル2（別のユーザー）
   ./test_ws_client -addr localhost:8080 -chat 1 -token "USER2_JWT_TOKEN"
   ```
   
   両方のクライアントでメッセージを送信すると、お互いにメッセージを受信できることを確認できます。

### 方法2: wscatを使用

1. **wscatをインストール**
   ```bash
   npm install -g wscat
   ```

2. **WebSocket接続**
   ```bash
   wscat -c "ws://localhost:8080/ws/chat/1" -H "Authorization: Bearer YOUR_JWT_TOKEN"
   ```

3. **メッセージを送信**
   接続後、以下のJSON形式でメッセージを送信:
   ```json
   {"content": "Hello, World!"}
   ```

### 方法3: ブラウザの開発者ツールを使用

1. ブラウザのコンソールで以下を実行:
   ```javascript
   const token = "YOUR_JWT_TOKEN";
   const chatID = 1;
   const ws = new WebSocket(`ws://localhost:8080/ws/chat/${chatID}`);
   
   ws.onopen = () => {
     console.log("Connected");
     ws.send(JSON.stringify({content: "Hello from browser!"}));
   };
   
   ws.onmessage = (event) => {
     console.log("Received:", event.data);
   };
   
   ws.onerror = (error) => {
     console.error("Error:", error);
   };
   ```

   **注意**: ブラウザから接続する場合、CORSの設定が必要な場合があります。

## 期待される動作

1. **接続時**: 
   - 接続が成功すると、サーバーからメッセージ履歴が送信されます
   - ログに "Connected" が表示されます

2. **メッセージ送信時**:
   - クライアントが送信したメッセージが同じチャットルーム内の他のクライアントにブロードキャストされます
   - メッセージはデータベースに保存されます

3. **メッセージ受信時**:
   - 他のクライアントから送信されたメッセージを受信します
   - メッセージ形式: `{"content":"...","chat_id":1,"sender_id":"..."}`

## トラブルシューティング

### 認証エラー
- JWTトークンが有効か確認
- トークンの形式が `Bearer <token>` であることを確認
- Auth0の設定（AUTH0_DOMAIN, AUTH0_AUDIENCE）が正しいか確認

### 接続エラー
- サーバーが起動しているか確認
- ポート8080が使用可能か確認
- チャットIDが存在するか確認

### メッセージが届かない
- 同じチャットIDで接続しているか確認
- データベース接続が正常か確認
- サーバーログを確認

## テストデータの準備

テスト前に、以下のデータがデータベースに存在することを確認:

1. **チャットルーム**: `chats`テーブルにテスト用のチャットレコード
2. **ユーザー**: テスト用のユーザーが`users`テーブルに存在
3. **メッセージ履歴**: 既存のメッセージがある場合、接続時に送信されます

