import { WS_BASE_URL } from '@/constants';
import { WebSocketMessage } from '@/types';
import { getAuthToken } from '@/utils/auth';
import useWebSocket from 'react-use-websocket';
import { useAuth } from './auth';

// https://tkdodo.eu/blog/using-web-sockets-with-react-query

export function useReactQuerySubscription() {
  const { accessToken } = useAuth();
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
  });
}
