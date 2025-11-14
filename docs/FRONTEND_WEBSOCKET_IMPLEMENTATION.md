# フロントエンド WebSocketチャット機能実装ガイド

このドキュメントは、MeetUP+RアプリケーションのフロントエンドでWebSocketチャット機能を実装するためのガイドです。

## 目次

- [概要](#概要)
- [前提条件](#前提条件)
- [WebSocket接続](#websocket接続)
- [認証](#認証)
- [メッセージの送受信](#メッセージの送受信)
- [エラーハンドリング](#エラーハンドリング)
- [実装例](#実装例)
- [ベストプラクティス](#ベストプラクティス)

---

## 概要

WebSocketチャット機能は、リアルタイムでメッセージの送受信を行うための機能です。

**エンドポイント**: `ws://localhost:8080/ws/chat/{chatId}`

**認証**: JWTトークンを`Authorization`ヘッダーに`Bearer`スキームで付与する必要があります。

---

## 前提条件

1. **JWTトークンの取得**
   - Auth0から取得した有効なJWTトークンが必要です
   - トークンは`Authorization: Bearer <TOKEN>`形式で送信します

2. **チャットIDの取得**
   - チャットルームのID（`chatId`）が必要です
   - `GET /api/v1/chats`エンドポイントで取得できます

3. **WebSocket対応ブラウザ**
   - モダンブラウザはすべて対応しています

---

## WebSocket接続

### 基本的な接続方法

```typescript
// TypeScript/JavaScript の例
const chatId = 1; // チャットルームID
const token = "your_jwt_token"; // JWTトークン

// WebSocket URLの構築
const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = `${wsProtocol}//localhost:8080/ws/chat/${chatId}`;

// WebSocket接続の確立
const ws = new WebSocket(wsUrl);

// 認証ヘッダーの設定
// 注意: ブラウザのWebSocket APIでは直接ヘッダーを設定できません
// そのため、接続前にHTTPリクエストでトークンを検証するか、
// サーバー側でクエリパラメータからトークンを取得する実装が必要です
```

### 認証の問題と解決策

ブラウザのWebSocket APIでは、接続時にHTTPヘッダーを直接設定できません。以下の解決策があります：

#### 解決策1: サーバー側でクエリパラメータからトークンを取得（推奨）

**サーバー側の実装が必要**:
```go
// サーバー側でクエリパラメータからトークンを取得する実装を追加
// 例: ws://localhost:8080/ws/chat/1?token=YOUR_JWT_TOKEN
```

**フロントエンド側**:
```typescript
const token = "your_jwt_token";
const chatId = 1;
const wsUrl = `ws://localhost:8080/ws/chat/${chatId}?token=${encodeURIComponent(token)}`;
const ws = new WebSocket(wsUrl);
```

#### 解決策2: 接続前にHTTPリクエストで認証（現在の実装）

現在の実装では、EchoのJWTミドルウェアがHTTPリクエストの`Authorization`ヘッダーを期待しています。ブラウザから直接接続する場合は、サーバー側の実装変更が必要です。

---

## 認証

### JWTトークンの取得

```typescript
// Auth0からトークンを取得する例
import { useAuth0 } from '@auth0/auth0-react'; // React用の例

const { getAccessTokenSilently } = useAuth0();

// トークンを取得
const token = await getAccessTokenSilently({
  audience: 'https://api.meetupr.com/',
  scope: 'read:messages write:messages'
});
```

### トークンの検証

接続前に、トークンが有効かどうかを確認することを推奨します：

```typescript
// トークンの有効性を確認
const verifyToken = async (token: string): Promise<boolean> => {
  try {
    const response = await fetch('http://localhost:8080/api/v1/users/me', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    return response.ok;
  } catch (error) {
    console.error('Token verification failed:', error);
    return false;
  }
};
```

---

## メッセージの送受信

### メッセージの送信

```typescript
// メッセージを送信
const sendMessage = (ws: WebSocket, content: string) => {
  if (ws.readyState === WebSocket.OPEN) {
    const message = {
      content: content
    };
    ws.send(JSON.stringify(message));
  } else {
    console.error('WebSocket is not open. Current state:', ws.readyState);
  }
};

// 使用例
sendMessage(ws, "こんにちは！");
```

### メッセージの受信

```typescript
// メッセージを受信
ws.onmessage = (event) => {
  try {
    const message = JSON.parse(event.data);
    console.log('Received message:', message);
    
    // メッセージの構造:
    // {
    //   "content": "メッセージ内容",
    //   "chat_id": 1,
    //   "sender_id": "auth0|xxxxxxxxxx"
    // }
    
    // UIにメッセージを表示する処理
    displayMessage(message);
  } catch (error) {
    console.error('Failed to parse message:', error);
  }
};
```

### メッセージ履歴の取得

接続時に、サーバーから自動的にメッセージ履歴が送信されます。受信したメッセージを順番に表示してください。

---

## エラーハンドリング

### 接続エラー

```typescript
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
  // エラー処理: ユーザーに通知、再接続の試行など
  showErrorMessage('接続エラーが発生しました');
};

ws.onclose = (event) => {
  console.log('WebSocket closed:', event.code, event.reason);
  
  // 正常な切断の場合
  if (event.code === 1000) {
    console.log('Connection closed normally');
  } else {
    // 異常な切断の場合、再接続を試行
    console.log('Connection closed unexpectedly, attempting to reconnect...');
    setTimeout(() => {
      connectWebSocket(chatId, token);
    }, 3000);
  }
};
```

### 認証エラー

```typescript
ws.onerror = (error) => {
  // サーバーから401エラーが返された場合
  if (error.target?.readyState === WebSocket.CLOSED) {
    // トークンの再取得を試行
    refreshTokenAndReconnect();
  }
};
```

### 再接続ロジック

```typescript
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
const reconnectDelay = 3000; // 3秒

const reconnect = (chatId: number, token: string) => {
  if (reconnectAttempts < maxReconnectAttempts) {
    reconnectAttempts++;
    setTimeout(() => {
      console.log(`Reconnection attempt ${reconnectAttempts}...`);
      connectWebSocket(chatId, token);
    }, reconnectDelay * reconnectAttempts);
  } else {
    console.error('Max reconnection attempts reached');
    showErrorMessage('接続に失敗しました。ページを再読み込みしてください。');
  }
};
```

---

## 実装例

### Vue.js/Nuxt.js の例

```vue
<template>
  <div class="chat-container">
    <div class="messages" ref="messagesContainer">
      <div
        v-for="message in messages"
        :key="message.id"
        :class="['message', { 'own-message': message.sender_id === currentUserId }]"
      >
        <div class="message-content">{{ message.content }}</div>
        <div class="message-time">{{ formatTime(message.sent_at) }}</div>
      </div>
    </div>
    <div class="input-area">
      <input
        v-model="inputMessage"
        @keyup.enter="sendMessage"
        placeholder="メッセージを入力..."
      />
      <button @click="sendMessage">送信</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue';

interface Message {
  id?: number;
  content: string;
  chat_id: number;
  sender_id: string;
  sent_at?: string;
}

const props = defineProps<{
  chatId: number;
  token: string;
  currentUserId: string;
}>();

const messages = ref<Message[]>([]);
const inputMessage = ref('');
const messagesContainer = ref<HTMLElement | null>(null);
let ws: WebSocket | null = null;

const connectWebSocket = () => {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${wsProtocol}//localhost:8080/ws/chat/${props.chatId}?token=${encodeURIComponent(props.token)}`;
  
  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log('WebSocket connected');
  };

  ws.onmessage = (event) => {
    try {
      const message: Message = JSON.parse(event.data);
      messages.value.push(message);
      scrollToBottom();
    } catch (error) {
      console.error('Failed to parse message:', error);
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  ws.onclose = (event) => {
    console.log('WebSocket closed:', event.code, event.reason);
    // 再接続ロジックを実装
  };
};

const sendMessage = () => {
  if (!inputMessage.value.trim() || !ws || ws.readyState !== WebSocket.OPEN) {
    return;
  }

  const message = {
    content: inputMessage.value.trim()
  };

  ws.send(JSON.stringify(message));
  inputMessage.value = '';
};

const scrollToBottom = async () => {
  await nextTick();
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  }
};

onMounted(() => {
  connectWebSocket();
});

onUnmounted(() => {
  if (ws) {
    ws.close(1000, 'Component unmounted');
  }
});
</script>
```

### React の例

```tsx
import { useEffect, useRef, useState } from 'react';

interface Message {
  id?: number;
  content: string;
  chat_id: number;
  sender_id: string;
  sent_at?: string;
}

interface ChatProps {
  chatId: number;
  token: string;
  currentUserId: string;
}

const Chat: React.FC<ChatProps> = ({ chatId, token, currentUserId }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const wsRef = useRef<WebSocket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//localhost:8080/ws/chat/${chatId}?token=${encodeURIComponent(token)}`;
    
    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onopen = () => {
      console.log('WebSocket connected');
    };

    wsRef.current.onmessage = (event) => {
      try {
        const message: Message = JSON.parse(event.data);
        setMessages((prev) => [...prev, message]);
      } catch (error) {
        console.error('Failed to parse message:', error);
      }
    };

    wsRef.current.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    wsRef.current.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
    };

    return () => {
      if (wsRef.current) {
        wsRef.current.close(1000, 'Component unmounted');
      }
    };
  }, [chatId, token]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const sendMessage = () => {
    if (!inputMessage.trim() || !wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      return;
    }

    const message = {
      content: inputMessage.trim()
    };

    wsRef.current.send(JSON.stringify(message));
    setInputMessage('');
  };

  return (
    <div className="chat-container">
      <div className="messages">
        {messages.map((message, index) => (
          <div
            key={index}
            className={`message ${message.sender_id === currentUserId ? 'own-message' : ''}`}
          >
            <div className="message-content">{message.content}</div>
            {message.sent_at && (
              <div className="message-time">{formatTime(message.sent_at)}</div>
            )}
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>
      <div className="input-area">
        <input
          value={inputMessage}
          onChange={(e) => setInputMessage(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
          placeholder="メッセージを入力..."
        />
        <button onClick={sendMessage}>送信</button>
      </div>
    </div>
  );
};
```

---

## ベストプラクティス

### 1. 接続状態の管理

```typescript
enum WebSocketState {
  CONNECTING = 0,
  OPEN = 1,
  CLOSING = 2,
  CLOSED = 3
}

const getConnectionState = (ws: WebSocket): string => {
  switch (ws.readyState) {
    case WebSocketState.CONNECTING:
      return '接続中...';
    case WebSocketState.OPEN:
      return '接続済み';
    case WebSocketState.CLOSING:
      return '切断中...';
    case WebSocketState.CLOSED:
      return '切断済み';
    default:
      return '不明';
  }
};
```

### 2. メッセージのキューイング

接続が確立される前にメッセージを送信しようとした場合、キューに保存して接続後に送信します：

```typescript
const messageQueue: string[] = [];

const queueMessage = (message: string) => {
  messageQueue.push(message);
};

const flushMessageQueue = (ws: WebSocket) => {
  while (messageQueue.length > 0 && ws.readyState === WebSocket.OPEN) {
    const message = messageQueue.shift();
    if (message) {
      ws.send(message);
    }
  }
};

ws.onopen = () => {
  console.log('WebSocket connected');
  flushMessageQueue(ws);
};
```

### 3. ハートビート（Ping/Pong）

サーバー側でPing/Pongが実装されているため、クライアント側では特別な処理は不要です。

### 4. メッセージの重複防止

```typescript
const receivedMessageIds = new Set<number>();

const handleMessage = (message: Message) => {
  if (message.id && receivedMessageIds.has(message.id)) {
    // 重複メッセージを無視
    return;
  }
  
  if (message.id) {
    receivedMessageIds.add(message.id);
  }
  
  // メッセージを処理
  displayMessage(message);
};
```

### 5. オフライン時の処理

```typescript
// オフライン検知
window.addEventListener('online', () => {
  console.log('Network is online, reconnecting...');
  reconnect(chatId, token);
});

window.addEventListener('offline', () => {
  console.log('Network is offline');
  showErrorMessage('ネットワーク接続が切断されました');
});
```

---

## API仕様の参照

詳細なAPI仕様については、以下を参照してください：

- [API仕様書](./API_SPECIFICATION.md#4-チャット-chats-ws)
- [データベース設計書](./DATABASE.md)

---

## トラブルシューティング

### 接続できない

1. **トークンが有効か確認**
   ```typescript
   const verifyToken = async (token: string) => {
     const response = await fetch('/api/v1/users/me', {
       headers: { 'Authorization': `Bearer ${token}` }
     });
     return response.ok;
   };
   ```

2. **サーバーが起動しているか確認**
   ```typescript
   const checkServer = async () => {
     try {
       const response = await fetch('http://localhost:8080/');
       return response.ok;
     } catch (error) {
       return false;
     }
   };
   ```

### メッセージが受信できない

1. **同じチャットIDで接続しているか確認**
2. **サーバーログを確認**（開発環境）
3. **ネットワークタブでWebSocket接続を確認**（ブラウザの開発者ツール）

### メッセージが送信できない

1. **WebSocketの状態を確認**
   ```typescript
   if (ws.readyState !== WebSocket.OPEN) {
     console.error('WebSocket is not open');
     return;
   }
   ```

2. **メッセージの形式を確認**
   ```typescript
   const message = { content: "メッセージ内容" };
   ws.send(JSON.stringify(message));
   ```

---

## セキュリティに関する注意事項

1. **トークンの保護**
   - トークンはローカルストレージやセッションストレージに保存しない
   - メモリ内でのみ管理することを推奨

2. **HTTPS/WSSの使用**
   - 本番環境では必ず`wss://`を使用
   - 開発環境でも可能な限り`wss://`を使用

3. **入力値の検証**
   - クライアント側でもメッセージ内容を検証
   - XSS対策を実装

---

## 参考リンク

- [WebSocket API (MDN)](https://developer.mozilla.org/ja/docs/Web/API/WebSocket)
- [Auth0 React SDK](https://auth0.com/docs/libraries/auth0-react)
- [Vue.js 公式ドキュメント](https://ja.vuejs.org/)
- [React 公式ドキュメント](https://ja.react.dev/)

