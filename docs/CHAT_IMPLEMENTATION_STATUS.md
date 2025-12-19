# チャット機能 実装状況まとめ

このドキュメントは、MeetUP+Rアプリケーションのチャット機能の実装状況を整理したものです。

最終更新: 2025-01-27

---

## 📋 概要

チャット機能は、ユーザー間のリアルタイムメッセージングを実現するための機能です。
WebSocketを使用したリアルタイム通信と、REST APIによるメッセージ履歴取得が実装されています。

**実装状況**: ✅ **実装済み**

---

## 🏗️ アーキテクチャ

### 技術スタック

- **リアルタイム通信**: WebSocket (gorilla/websocket)
- **REST API**: Echo Framework
- **データベース**: Supabase (PostgreSQL)
- **認証**: Auth0 JWT

### コンポーネント構成

```
┌─────────────────┐
│   Frontend      │
│  (WebSocket)    │
└────────┬────────┘
         │
         │ WS /ws/chat/{chatId}
         │
┌────────▼─────────────────────────┐
│   Echo Server                    │
│  ┌─────────────────────────────┐  │
│  │  WebSocket Handler         │  │
│  │  - WsHandler               │  │
│  │  - Hub (Room Management)   │  │
│  │  - Client (Connection)     │  │
│  └─────────────────────────────┘  │
│  ┌─────────────────────────────┐  │
│  │  REST API Handlers         │  │
│  │  - GetChats                │  │
│  │  - GetChatMessages         │  │
│  └─────────────────────────────┘  │
└────────┬──────────────────────────┘
         │
         │ Supabase API
         │
┌────────▼────────┐
│   Supabase      │
│  ┌───────────┐  │
│  │  chats    │  │
│  │  messages │  │
│  └───────────┘  │
└─────────────────┘
```

---

## 📡 API エンドポイント

### 1. REST API

#### `GET /api/v1/chats`
- **説明**: ユーザーが参加しているチャットルームの一覧を取得
- **認証**: 必要 (Auth0 JWT)
- **ハンドラー**: `handlers.GetChats`
- **実装ファイル**: `internal/handlers/chat.go`
- **データベース関数**: `db.GetUserChats`
- **レスポンス**: 
  - チャットルームのリスト（`models.Chat`配列）
  - 各チャットには相手ユーザー情報（`OtherUser`）と最終メッセージ（`LastMessage`）が含まれる
- **ソート**: 作成日時降順（最新が先頭）

#### `GET /api/v1/chats/{chatId}/messages`
- **説明**: 特定のチャットルームのメッセージ履歴を取得
- **認証**: 必要 (Auth0 JWT)
- **ハンドラー**: `handlers.GetChatMessages`
- **実装ファイル**: `internal/handlers/chat.go`
- **データベース関数**: `db.GetChatMessages`
- **セキュリティ**: 
  - チャット参加者のみアクセス可能（`db.IsChatParticipant`で検証）
- **レスポンス**: 
  - メッセージのリスト（`models.Message`配列）
- **ソート**: 送信日時昇順（古い順）

### 2. WebSocket API

#### `WS /ws/chat/{chatId}`
- **説明**: WebSocketを使用したリアルタイムメッセージ送受信
- **認証**: 必要 (Auth0 JWT via Echo middleware)
- **ハンドラー**: `handlers.WsHandler`
- **実装ファイル**: `internal/handlers/websocket.go`
- **接続管理**: 
  - Hubによるルーム管理（`chatID`ごとにルーム分離）
  - クライアント接続の自動登録・解除
- **機能**:
  - リアルタイムメッセージ送信
  - リアルタイムメッセージ受信（ブロードキャスト）
  - 接続時のメッセージ履歴自動送信
  - メッセージの自動データベース保存

---

## 💾 データモデル

### Chat (チャットルーム)

```go
type Chat struct {
    ID               int64     `json:"id"`
    User1ID          string    `json:"user1_id"`
    User2ID          string    `json:"user2_id"`
    AISuggestedTheme string    `json:"ai_suggested_theme,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
    OtherUser        *User     `json:"other_user,omitempty"`
    LastMessage      *Message  `json:"last_message,omitempty"`
}
```

**データベーステーブル**: `chats`
- 1対1チャットのみ対応（`user1_id`と`user2_id`）
- ユニーク制約: 同じユーザーペアの重複チャットを防止
- AI提案テーマフィールドあり（未使用）

### Message (メッセージ)

```go
type Message struct {
    ID                int64     `json:"id"`
    ChatID            int64     `json:"chat_id"`
    SenderID          string    `json:"sender_id"`
    Content           string    `json:"content"`
    TranslatedContent string    `json:"translated_content,omitempty"`
    MessageType       string    `json:"message_type"`
    SentAt            time.Time `json:"sent_at"`
}
```

**データベーステーブル**: `messages`
- 現在は`message_type`は常に`"text"`として保存
- 翻訳機能用の`translated_content`フィールドあり（未使用）

---

## 🔧 実装詳細

### WebSocket Hub

**ファイル**: `internal/handlers/websocket.go`

**構造**:
- `Hub`: 全クライアントとルームを管理
  - `clients`: 全接続クライアント
  - `rooms`: `chatID`ごとのクライアントマップ
  - `broadcast`: メッセージブロードキャストチャネル
  - `register`: クライアント登録チャネル
  - `unregister`: クライアント解除チャネル

**動作フロー**:
1. クライアント接続 → `WsHandler`が呼ばれる
2. WebSocket接続をアップグレード
3. `Client`構造体を作成し、Hubに登録
4. メッセージ履歴を非同期で読み込み・送信
5. `readPump`と`writePump`をgoroutineで起動
6. メッセージ受信 → データベースに保存 → 同じルームの全クライアントにブロードキャスト

**設定値**:
- `writeWait`: 10秒
- `pongWait`: 60秒
- `pingPeriod`: 54秒（pongWaitの90%）
- `maxMessageSize`: 512バイト

### データベース操作

**ファイル**: `internal/db/db.go`

**主要関数**:

1. **`GetUserChats(userID string)`**
   - ユーザーが参加している全チャットを取得
   - `user1_id`と`user2_id`の両方で検索
   - 重複排除（同じチャットID）
   - 相手ユーザー情報と最終メッセージを付与
   - 作成日時降順でソート

2. **`GetChatMessages(chatID int64)`**
   - チャットルームの全メッセージを取得
   - 送信日時昇順でソート

3. **`GetLastChatMessage(chatID int64)`**
   - チャットルームの最新メッセージを1件取得
   - `GetUserChats`で使用

4. **`IsChatParticipant(chatID int64, userID string)`**
   - ユーザーがチャットの参加者かどうかを検証
   - `GetChatMessages`の認可チェックで使用

---

## 🔐 セキュリティ

### 認証

- **REST API**: EchoのJWTミドルウェア（`auth.EchoJWTMiddleware()`）で保護
- **WebSocket**: 同じJWTミドルウェアを使用（接続時に検証）
- **認可**: 
  - `GetChatMessages`: チャット参加者のみアクセス可能
  - WebSocket: 接続時にJWTから`user_id`を取得し、チャット参加者かどうかを検証する必要あり（現在は実装されていない可能性）

### データベース

- SupabaseのRow Level Security (RLS) ポリシーに依存
- 外部キー制約により、存在しないユーザーやチャットへの参照を防止

---

## ⚠️ 制限事項・課題

### 現在の制限

1. **メッセージサイズ制限**
   - `maxMessageSize = 512`バイト（非常に小さい）
   - 長文メッセージが送信できない

2. **メッセージタイプ**
   - 現在は`"text"`のみ対応
   - 画像、スタンプなどのメディアタイプ未対応

3. **翻訳機能**
   - `translated_content`フィールドは存在するが未使用
   - リアルタイム翻訳機能未実装

4. **WebSocket認証**
   - ブラウザからの直接接続時の認証方法が不明確
   - クエリパラメータでのトークン受け渡しが推奨されているが、実装状況不明

5. **エラーハンドリング**
   - データベースエラー時の詳細なエラーレスポンスが不足
   - クライアント側へのエラー通知が限定的

6. **パフォーマンス**
   - メッセージ履歴の取得時に全件取得（ページネーションなし）
   - 大量のメッセージがある場合のパフォーマンス懸念

### 未実装機能

1. **チャット作成API**
   - チャットルームを手動で作成するAPIが存在しない
   - 現在は「匿名会いたいボタン」の相互マッチ時に自動作成される想定（未実装）ゆ

2. **メッセージ削除・編集**
   - メッセージの削除・編集機能なし

3. **既読機能**
   - メッセージの既読状態管理なし

4. **通知機能**
   - 新着メッセージの通知機能なし

5. **AIテーマ提案**
   - `ai_suggested_theme`フィールドは存在するが、設定・取得APIなし

---

## 📝 テスト

### テストファイル

- `internal/handlers/websocket_test.go`: WebSocketハンドラーのテスト
- `test_websocket_client.go`: WebSocketクライアントのテストコード
- `test_websocket.html`: ブラウザでのWebSocketテスト用HTML
- `test_websocket.sh`: WebSocket接続テストスクリプト

### テストデータ作成

- `scripts/create_test_chat_data.go`: テスト用チャットデータ作成
- `scripts/create_test_messages.go`: テスト用メッセージデータ作成
- `scripts/check_chats.go`: チャットデータ確認スクリプト

---

## 📚 関連ドキュメント

- [API仕様書](./API_SPECIFICATION.md) - チャットAPIの詳細仕様
- [データベース設計書](./DATABASE.md) - チャット関連テーブルのスキーマ
- [フロントエンド実装ガイド](./FRONTEND_WEBSOCKET_IMPLEMENTATION.md) - フロントエンドでの実装方法
- [実装状況まとめ](./IMPLEMENTATION_STATUS.md) - 全体の実装状況

---

## 🎯 今後の改善提案

### 高優先度

1. **メッセージサイズ制限の拡大**
   - `maxMessageSize`を512バイトから10KB程度に拡大

2. **WebSocket認証の明確化**
   - クエリパラメータでのトークン受け渡しを実装
   - または、接続前のHTTP認証エンドポイントを提供

3. **チャット作成APIの実装**
   - `POST /api/v1/chats`エンドポイントの追加
   - または、「匿名会いたいボタン」機能の実装

### 中優先度

4. **ページネーション対応**
   - メッセージ履歴取得時のページネーション実装

5. **エラーハンドリングの改善**
   - より詳細なエラーレスポンス
   - クライアント側への適切なエラー通知

6. **既読機能の実装**
   - メッセージ既読状態の管理

### 低優先度

7. **メディアメッセージ対応**
   - 画像、スタンプなどのメッセージタイプ追加

8. **翻訳機能の実装**
   - リアルタイム翻訳機能

9. **通知機能の実装**
   - プッシュ通知やWebSocket経由の通知

---

## 📊 実装状況サマリー

| 機能 | 実装状況 | 備考 |
|------|---------|------|
| チャット一覧取得 | ✅ 実装済み | `GET /api/v1/chats` |
| メッセージ履歴取得 | ✅ 実装済み | `GET /api/v1/chats/{chatId}/messages` |
| リアルタイムメッセージ送信 | ✅ 実装済み | WebSocket |
| リアルタイムメッセージ受信 | ✅ 実装済み | WebSocket |
| メッセージ自動保存 | ✅ 実装済み | データベースに自動保存 |
| メッセージ履歴自動送信 | ✅ 実装済み | 接続時に自動送信 |
| チャット作成API | ❌ 未実装 | 手動作成APIなし |
| メッセージ削除・編集 | ❌ 未実装 | - |
| 既読機能 | ❌ 未実装 | - |
| 翻訳機能 | ❌ 未実装 | フィールドは存在 |
| メディアメッセージ | ❌ 未実装 | テキストのみ |
| AIテーマ提案 | ⚠️ 部分的 | フィールドは存在、APIなし |

---

最終更新: 2025-01-27
