# Feat: ユーザー機能の新規実装

## 概要

本プルリクエストは、アプリケーションの根幹機能であるユーザー管理機能を追加するものです。Auth0による認証後のユーザー登録、プロフィール情報の取得・更新、および他ユーザーの検索機能を提供します。

## 実装内容

### 1. APIエンドポイントの追加 (`cmd/meetupr-backend/main.go`)

`echo`フレームワークを使用し、`/api/v1/users` グループ配下に以下のエンドポイントを定義しました。全てのエンドポイントはJWTミドルウェアによる認証が必要です。

- `POST /register`: 新規ユーザー登録
- `GET /me`: 自身のプロフィール詳細を取得
- `PUT /me`: 自身のプロフィールを更新
- `GET /`: 条件を指定してユーザーを検索
- `GET /:userId`: 特定のユーザーの公開プロフィールを取得

### 2. データモデルの定義 (`internal/models/user.go`)

ユーザーおよび関連情報のデータ構造を定義しました。

- **`User`**: ユーザーの基本情報とプロフィール情報を統合した主要モデル。
- **`Interest`**: ユーザーの興味・関心事を表すモデル。
- **`UserProfileResponse`**: プロフィール取得時のレスポンス専用モデル。
- **`RegisterUserRequest`**: ユーザー登録APIのリクエストボディ用モデル。
- **`UpdateUserProfileRequest`**: プロフィール更新APIのリクエストボディ用モデル。

### 3. データベース層の実装 (`internal/db/db.go`)

Supabase Goライブラリを利用して、データベースとのインタラクションを行うロジックを実装しました。

- **`CreateUser`**: `users`テーブルにレコードを挿入後、関連する`profiles`テーブルにもデフォルトレコードを作成します。
- **`GetUserByID`**: ユーザーIDを基に、`profiles`と`user_interests`をJOINして包括的なプロフィール情報を取得します。
- **`UpdateUserProfile`**: `users`テーブルの`username`と、`profiles`テーブルの各種情報を更新します。`user_interests`は一度削除してから再登録する方式を採っています。
- **`SearchUsers`**: 興味ID、学習言語、話せる言語を条件にユーザーを絞り込み検索します。
- **`GetUserProfile`**: 他のユーザーの公開用プロフィール情報を取得します。

### 4. ハンドラ層の実装 (`internal/handlers/user.go`)

APIエンドポイントの具体的な処理ロジックを実装しました。

- JWTトークンの`context`から認証済みユーザーのID (`user_id`) とメールアドレス (`user_email`) を取得し、操作の主体を特定します。
- リクエストボディのバリデーションを行い、不正なリクエストは`400 Bad Request`を返します。
- データベース層の関数を呼び出し、結果に応じて成功レスポンス (`200 OK` or `201 Created`) またはエラーレスポンス (`404 Not Found`, `500 Internal Server Error`など) を返却します。

## 技術的詳細

- **関連テーブルのJOIN**: Supabaseの`Select`メソッドで`profiles(*), user_interests(*, interests(*))`のように記述することで、Go側で効率的に関連データを取得しています。
- **エラーハンドリング**: `db`層から返されるエラーをハンドラ層で適切に解釈し、`echo.HTTPError`を用いてクライアントにステータスコードとメッセージを返却するよう統一しています。

## 確認事項

- [ ] `docs/API_SPECIFICATION.md` に記載された仕様と、実装されたAPIの挙動が一致しているか。
- [ ] データベースのマイグレーションは完了しているか。（`users`, `profiles`, `interests`, `user_interests`テーブルが必要）
- [ ] `POST /register` 実行時に、`users`と`profiles`の両テーブルにレコードが作成されるか。
- [ ] `PUT /me` で興味・関心を変更した際、`user_interests`テーブルが正しく更新されるか。
- [ ] 各エンドポイントで、予期せぬエラーが発生した場合に適切なHTTPステータスコードが返されるか。