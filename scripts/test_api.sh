#!/bin/bash

# テスト用APIリクエストスクリプト
# DISABLE_AUTH=true モードでサーバーを起動している場合に使用

# デフォルトのテストユーザーID
TEST_USER_ID="${1:-auth0|6917784d99703fe24aebd01d}"
TEST_USER_EMAIL="${2:-testuser1@example.com}"

echo "🔧 Testing with User ID: $TEST_USER_ID"
echo "📧 Email: $TEST_USER_EMAIL"
echo ""

# チャット一覧を取得
echo "📋 Getting chat list..."
curl -X GET http://localhost:8080/api/v1/chats \
  -H "X-Test-User-ID: $TEST_USER_ID" \
  -H "X-Test-User-Email: $TEST_USER_EMAIL" \
  -H "Content-Type: application/json" \
  | jq '.'

echo ""
echo ""

# チャットメッセージを取得（チャットIDを指定）
if [ -n "$3" ]; then
  CHAT_ID="$3"
  echo "💬 Getting messages for chat $CHAT_ID..."
  curl -X GET "http://localhost:8080/api/v1/chats/$CHAT_ID/messages" \
    -H "X-Test-User-ID: $TEST_USER_ID" \
    -H "X-Test-User-Email: $TEST_USER_EMAIL" \
    -H "Content-Type: application/json" \
    | jq '.'
fi

