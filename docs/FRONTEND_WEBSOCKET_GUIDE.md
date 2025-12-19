# フロントエンド WebSocket 実装ガイド

## 📋 概要

このガイドでは、バックエンドのWebSocket APIを使用してチャット機能を実装する方法を説明します。

## 🔌 WebSocket接続

### エンドポイント
```
ws://localhost:8080/ws/chat/{chatID}?token={JWT_TOKEN}
```

### 接続例（TypeScript/Vue.js）

```typescript
// composables/useChatWebSocket.ts
import { ref, onUnmounted } from 'vue'

interface Message {
  id: number
  chat_id: number
  sender_id: string
  content: string
  message_type: string
  sent_at: string // ISO 8601形式の文字列
}

export function useChatWebSocket(chatId: number, token: string) {
  const messages = ref<Message[]>([])
  const isConnected = ref(false)
  const ws = ref<WebSocket | null>(null)

  const connect = () => {
    const wsUrl = `ws://localhost:8080/ws/chat/${chatId}?token=${token}`
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      console.log('WebSocket接続が確立されました')
      isConnected.value = true
    }

    ws.value.onmessage = (event) => {
      try {
        // バックエンドから送信される各メッセージをパース
        const message: Message = JSON.parse(event.data)
        console.log('メッセージを受信:', message)
        
        // メッセージをリストに追加
        messages.value.push(message)
        
        // 必要に応じて、メッセージを送信日時でソート
        messages.value.sort((a, b) => 
          new Date(a.sent_at).getTime() - new Date(b.sent_at).getTime()
        )
      } catch (error) {
        console.error('メッセージのパースエラー:', error)
      }
    }

    ws.value.onerror = (error) => {
      console.error('WebSocketエラー:', error)
      isConnected.value = false
    }

    ws.value.onclose = () => {
      console.log('WebSocket接続が閉じられました')
      isConnected.value = false
    }
  }

  const sendMessage = (content: string) => {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      // バックエンドは content フィールドのみを期待
      const message = {
        content: content
      }
      ws.value.send(JSON.stringify(message))
    } else {
      console.error('WebSocketが接続されていません')
    }
  }

  const disconnect = () => {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
  }

  // コンポーネントがアンマウントされたときに接続を閉じる
  onUnmounted(() => {
    disconnect()
  })

  return {
    messages,
    isConnected,
    connect,
    sendMessage,
    disconnect
  }
}
```

## 📨 メッセージ形式

### 受信メッセージ（バックエンド → フロントエンド）

バックエンドから送信されるメッセージは、以下の形式です：

```typescript
interface Message {
  id: number              // メッセージID（データベースのID）
  chat_id: number        // チャットID
  sender_id: string      // 送信者のユーザーID（Auth0のsub）
  content: string        // メッセージ内容
  message_type: string    // メッセージタイプ（通常は "text"）
  sent_at: string        // 送信日時（ISO 8601形式、例: "2025-12-19T01:33:17Z"）
}
```

**重要**: 
- 接続時に、バックエンドが自動的にメッセージ履歴を送信します
- 各メッセージは**個別のWebSocketメッセージ**として送信されます（改行区切りではありません）
- 履歴メッセージとリアルタイムメッセージは同じ形式です

### 送信メッセージ（フロントエンド → バックエンド）

フロントエンドから送信するメッセージは、以下の形式です：

```typescript
{
  content: string  // メッセージ内容のみ
}
```

**注意**: `chat_id` と `sender_id` は、WebSocket接続時に既に確定しているため、送信する必要はありません。

## 🎯 Vue.jsコンポーネントでの使用例

```vue
<template>
  <div class="chat-container">
    <div class="messages">
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
        v-model="newMessage" 
        @keyup.enter="handleSend"
        placeholder="メッセージを入力..."
      />
      <button @click="handleSend" :disabled="!isConnected">
        送信
      </button>
    </div>
    
    <div v-if="!isConnected" class="connection-status">
      接続中...
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useChatWebSocket } from '@/composables/useChatWebSocket'
import { useAuth0 } from '@auth0/auth0-vue'

const props = defineProps<{
  chatId: number
}>()

const auth0 = useAuth0()
const newMessage = ref('')

// WebSocket接続を確立
const { messages, isConnected, connect, sendMessage, disconnect } = useChatWebSocket(
  props.chatId,
  auth0.getAccessTokenSilently() // または適切な方法でトークンを取得
)

const currentUserId = auth0.user.value?.sub

const handleSend = () => {
  if (newMessage.value.trim() && isConnected.value) {
    sendMessage(newMessage.value.trim())
    newMessage.value = ''
  }
}

const formatTime = (sentAt: string) => {
  const date = new Date(sentAt)
  return date.toLocaleTimeString('ja-JP', { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

onMounted(() => {
  // コンポーネントがマウントされたときに接続
  connect()
})

onUnmounted(() => {
  // コンポーネントがアンマウントされたときに切断
  disconnect()
})
</script>
```

## 🔍 デバッグのポイント

### 1. 接続確認

ブラウザの開発者ツール（F12）のコンソールで以下を確認：

```javascript
// WebSocket接続が確立されているか確認
console.log('WebSocket状態:', ws.readyState)
// readyState: 0=CONNECTING, 1=OPEN, 2=CLOSING, 3=CLOSED
```

### 2. メッセージ受信確認

`onmessage` イベントハンドラで受信したメッセージをログ出力：

```typescript
ws.value.onmessage = (event) => {
  console.log('受信した生データ:', event.data)
  const message = JSON.parse(event.data)
  console.log('パース後のメッセージ:', message)
  // ...
}
```

### 3. エラーハンドリング

```typescript
ws.value.onerror = (error) => {
  console.error('WebSocketエラー:', error)
  // エラー時の処理（再接続など）
}

ws.value.onclose = (event) => {
  console.log('接続が閉じられました:', event.code, event.reason)
  // 必要に応じて再接続ロジックを実装
}
```

## ⚠️ 注意事項

1. **認証トークン**: JWTトークンはクエリパラメータ `?token=...` として渡す必要があります（WebSocketではカスタムヘッダーが設定できないため）

2. **メッセージの重複**: 履歴メッセージとリアルタイムメッセージで重複が発生する可能性があります。`id` フィールドで重複チェックを行ってください：

```typescript
const messageIds = new Set<number>()

ws.value.onmessage = (event) => {
  const message: Message = JSON.parse(event.data)
  
  // 重複チェック
  if (!messageIds.has(message.id)) {
    messageIds.add(message.id)
    messages.value.push(message)
  }
}
```

3. **接続のライフサイクル**: コンポーネントがアンマウントされたときや、ページを離れるときに必ずWebSocket接続を閉じてください

4. **再接続**: ネットワークエラーなどで接続が切れた場合の再接続ロジックを実装することを推奨します

## 📝 完全な実装例（TypeScript）

```typescript
// composables/useChatWebSocket.ts
import { ref, onUnmounted, Ref } from 'vue'

interface Message {
  id: number
  chat_id: number
  sender_id: string
  content: string
  message_type: string
  sent_at: string
}

export function useChatWebSocket(chatId: Ref<number> | number, token: Ref<string> | string) {
  const messages = ref<Message[]>([])
  const isConnected = ref(false)
  const ws = ref<WebSocket | null>(null)
  const messageIds = new Set<number>()

  const getChatId = () => typeof chatId === 'number' ? chatId : chatId.value
  const getToken = () => typeof token === 'string' ? token : token.value

  const connect = () => {
    if (ws.value?.readyState === WebSocket.OPEN) {
      console.log('既に接続されています')
      return
    }

    const wsUrl = `ws://localhost:8080/ws/chat/${getChatId()}?token=${getToken()}`
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      console.log('WebSocket接続が確立されました')
      isConnected.value = true
      messageIds.clear() // 再接続時は重複チェックをリセット
    }

    ws.value.onmessage = (event) => {
      try {
        const message: Message = JSON.parse(event.data)
        
        // 重複チェック
        if (messageIds.has(message.id)) {
          console.log('重複メッセージをスキップ:', message.id)
          return
        }
        
        messageIds.add(message.id)
        messages.value.push(message)
        
        // 送信日時でソート
        messages.value.sort((a, b) => 
          new Date(a.sent_at).getTime() - new Date(b.sent_at).getTime()
        )
        
        console.log('メッセージを受信:', message)
      } catch (error) {
        console.error('メッセージのパースエラー:', error, event.data)
      }
    }

    ws.value.onerror = (error) => {
      console.error('WebSocketエラー:', error)
      isConnected.value = false
    }

    ws.value.onclose = (event) => {
      console.log('WebSocket接続が閉じられました:', event.code, event.reason)
      isConnected.value = false
      
      // 異常終了の場合は再接続を試みる（オプション）
      if (event.code !== 1000) { // 1000 = 正常終了
        console.log('再接続を試みます...')
        setTimeout(() => connect(), 3000)
      }
    }
  }

  const sendMessage = (content: string) => {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      const message = { content }
      ws.value.send(JSON.stringify(message))
      console.log('メッセージを送信:', message)
    } else {
      console.error('WebSocketが接続されていません')
      throw new Error('WebSocket接続が確立されていません')
    }
  }

  const disconnect = () => {
    if (ws.value) {
      ws.value.close(1000, '正常終了') // 1000 = 正常終了コード
      ws.value = null
      isConnected.value = false
    }
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    messages,
    isConnected,
    connect,
    sendMessage,
    disconnect
  }
}
```

## 🚀 次のステップ

1. 上記のコードをフロントエンドプロジェクトに実装
2. ブラウザの開発者ツールでWebSocket接続とメッセージの送受信を確認
3. 必要に応じてUIの改善（ローディング状態、エラー表示など）

