# API仕様書

このドキュメントは、MeetUP+RアプリケーションのバックエンドAPIの仕様を定義します。

## 認証

認証にはAuth0を使用します。
APIリクエストには、Auth0から発行されたJWT（JSON Web Token）を`Authorization`ヘッダーに`Bearer`スキームを用いて付与する必要があります。

**ヘッダー例:**
`Authorization: Bearer <YOUR_AUTH0_JWT>`

保護されたエンドポイントは、リクエストに有効なJWTが含まれていない場合、`401 Unauthorized`エラーを返します。

---

## エンドポイント一覧

### 1. ユーザー (`/users`)

#### `POST /api/v1/users/register`

-   **説明:** Auth0での初回認証後、ユーザー情報をアプリケーションのデータベースに登録します。
-   **認証:** 必要 (Auth0 JWT)
-   **リクエストボディ:**
    ```json
    {
      "username": "your_username"
    }
    ```
-   **レスポンス:**
    -   `201 Created`: 登録成功
        ```json
        {
          "id": "auth0|xxxxxxxxxx",
          "email": "user@example.com",
          "username": "your_username",
          "is_oic_verified": false,
          "created_at": "2025-11-14T10:00:00Z"
        }
        ```
    -   `409 Conflict`: ユーザーが既に存在する場合

#### `GET /api/v1/users/me`

-   **説明:** 現在ログインしているユーザー自身の詳細なプロフィール情報を取得します。
-   **認証:** 必要
-   **レスポンス:**
    -   `200 OK`:
        ```json
        {
          "user_id": "auth0|xxxxxxxxxx",
          "email": "user@example.com",
          "username": "your_username",
          "major": "情報理工学部",
          "gender": "男性",
          "native_language": "日本語",
          "spoken_languages": ["英語"],
          "learning_languages": ["韓国語"],
          "residence": "大阪",
          "comment": "よろしくお願いします！",
          "interests": [
            { "id": 1, "name": "ゲーム", "preference_level": 5 },
            { "id": 2, "name": "K-POP", "preference_level": 4 }
          ],
          "last_updated": "2025-11-14T11:00:00Z"
        }
        ```

#### `PUT /api/v1/users/me`

-   **説明:** 現在ログインしているユーザー自身のプロフィール情報を更新します。
-   **認証:** 必要
-   **リクエストボディ:**
    ```json
    {
      "username": "new_username",
      "major": "情報理工学部",
      "gender": "男性",
      "native_language": "日本語",
      "spoken_languages": ["英語"],
      "learning_languages": ["韓国語"],
      "residence": "大阪",
      "comment": "よろしくお願いします！",
      "interest_ids": [1, 2]
    }
    ```
-   **レスポンス:**
    -   `200 OK`: 更新後のプロフィール情報

#### `GET /api/v1/users`

-   **説明:** 条件に基づいて他のユーザーを検索します。
-   **認証:** 必要
-   **クエリパラメータ:**
    -   `interest_id` (number): 興味ID
    -   `learning_language` (string): 学習したい言語
    -   `spoken_language` (string): 話せる言語
-   **レスポンス:**
    -   `200 OK`: ユーザーのリスト

#### `GET /api/v1/users/{userId}`

-   **説明:** 特定のユーザーの公開プロフィール情報を取得します。
-   **認証:** 必要
-   **レスポンス:**
    -   `200 OK`: ユーザーの公開プロフィール

---

### 2. 興味・趣味 (`/interests`)

#### `GET /api/v1/interests`

-   **説明:** 登録可能な興味・趣味のマスターデータを一覧で取得します。
-   **認証:** 必要
-   **レスポンス:**
    -   `200 OK`:
        ```json
        [
          { "id": 1, "name": "ゲーム", "category": "インドア" },
          { "id": 2, "name": "K-POP", "category": "音楽" }
        ]
        ```

---

### 3. 匿名会いたいボタン (`/meet-requests`)

#### `POST /api/v1/meet-requests`

-   **説明:** 他のユーザーに「会いたい」リクエストを送信します。
-   **認証:** 必要
-   **リクエストボディ:**
    ```json
    {
      "recipient_id": "auth0|yyyyyyyyyy"
    }
    ```
-   **レスポンス:**
    -   `201 Created`: 作成成功
    -   `200 OK`: 相互マッチが成立した場合
        ```json
        {
          "match": true,
          "chat_id": 123
        }
        ```

---

### 4. チャット (`/chats`, `/ws`)

#### `GET /api/v1/chats`

-   **説明:** 自身が参加しているチャットルームの一覧を取得します。
-   **認証:** 必要
-   **レスポンス:**
    -   `200 OK`: チャットルームのリスト

#### `GET /api/v1/chats/{chatId}/messages`

-   **説明:** 特定のチャットルームのメッセージ履歴を取得します。
-   **認証:** 必要
-   **レスポンス:**
    -   `200 OK`: メッセージのリスト

#### `WS /ws/chat/{chatId}`

-   **説明:** WebSocketを使用してリアルタイムなメッセージ送受信を行います。
-   **認証:** 必要 (接続時にJWTを渡す)
-   **送信メッセージ:**
    ```json
    {
      "type": "text",
      "content": "こんにちは！"
    }
    ```
-   **受信メッセージ:**
    ```json
    {
      "id": 1,
      "chat_id": 123,
      "sender_id": "auth0|xxxxxxxxxx",
      "content": "こんにちは！",
      "translated_content": "Hello!",
      "message_type": "text",
      "sent_at": "2025-11-14T12:00:00Z"
    }
    ```

---

### 5. イベント (`/events`)

#### `GET /api/v1/events`

-   **説明:** 学校主催のイベントやミッションの一覧を取得します。
-   **認証:** 必要
-   **レスポンス:**
    -   `200 OK`: イベントのリスト

#### `GET /api/v1/events/{eventId}`

-   **説明:** 特定のイベントの詳細を取得します。
-   **認証:** 必要
-   **レスポンス:**
    -   `200 OK`: イベント詳細
