#!/bin/bash

# WebSocketチャット機能のテストスクリプト

echo "=========================================="
echo "WebSocketチャット機能テスト"
echo "=========================================="
echo ""

# 色の定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# パラメータの確認
if [ -z "$1" ]; then
    echo -e "${RED}エラー: JWTトークンが必要です${NC}"
    echo "使用方法: ./test_websocket.sh <JWT_TOKEN> [CHAT_ID] [SERVER_ADDR]"
    echo "例: ./test_websocket.sh 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...' 1 localhost:8080"
    exit 1
fi

TOKEN=$1
CHAT_ID=${2:-1}
SERVER_ADDR=${3:-localhost:8080}

echo -e "${YELLOW}設定:${NC}"
echo "  サーバー: $SERVER_ADDR"
echo "  チャットID: $CHAT_ID"
echo "  トークン: ${TOKEN:0:20}..."
echo ""

# サーバーが起動しているか確認
echo -e "${YELLOW}サーバーの接続確認中...${NC}"
if ! curl -s -o /dev/null -w "%{http_code}" "http://$SERVER_ADDR/" | grep -q "200"; then
    echo -e "${RED}警告: サーバーに接続できません。サーバーが起動しているか確認してください。${NC}"
    echo "サーバー起動コマンド: go run cmd/meetupr-backend/main.go"
    read -p "続行しますか? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✓ サーバーに接続できました${NC}"
fi

echo ""
echo -e "${YELLOW}テストクライアントを起動します...${NC}"
echo "Ctrl+Cで終了します"
echo ""

# テストクライアントを実行
./test_ws_client -addr "$SERVER_ADDR" -chat "$CHAT_ID" -token "$TOKEN"

