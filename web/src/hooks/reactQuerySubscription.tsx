import { getPrivateConversationQueryKey } from '@/api/user';
import { WS_BASE_URL } from '@/constants';
import { MessageCreatedPayload, WebSocketMessage } from '@/types';
import { getAuthToken } from '@/utils/auth';
import { useQueryClient } from '@tanstack/react-query';
import useWebSocket from 'react-use-websocket';
import { useAuth } from './auth';

// https://tkdodo.eu/blog/using-web-sockets-with-react-query

export function useReactQuerySubscription() {
  const { user, accessToken } = useAuth();
  const queryClient = useQueryClient();
  const { sendJsonMessage } = useWebSocket<WebSocketMessage>(WS_BASE_URL, {
    share: true,
    shouldReconnect: () => !!getAuthToken(), // Prevent reconnection if no auth token.
    onOpen: () => {
      sendJsonMessage<WebSocketMessage>({
        type: 'authentication',
        payload: {
          access_token: accessToken,
        },
      });
    },
    onMessage: (event) => {
      const message: WebSocketMessage = JSON.parse(event.data);
      switch (message.type) {
        case 'message_created': {
          const payload = message.payload as MessageCreatedPayload;
          const queryKey = getPrivateConversationQueryKey(
            user.id,
            payload.recipient_id
          );
          queryClient.invalidateQueries({ queryKey: queryKey });
          break;
        }
        case 'message_updated': {
          break;
        }
        case 'message_deleted': {
          break;
        }
        default: {
          console.log(`Invalid websocket message type: ${message.type}.`);
        }
      }
    },
  });
}
