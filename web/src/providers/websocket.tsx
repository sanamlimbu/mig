import { WS_BASE_URL } from '@/constants';
import { WebSocketContext } from '@/contexts/websocket';
import { WebSocketMessage } from '@/types';
import { getAuthToken } from '@/utils/auth';
import { PropsWithChildren } from 'react';
import useWebSocket from 'react-use-websocket';

export default function WebsocketProvider(props: PropsWithChildren) {
  const { sendJsonMessage, lastJsonMessage, readyState } =
    useWebSocket<WebSocketMessage>(WS_BASE_URL, {
      share: true,
      shouldReconnect: () => !!getAuthToken(), // Prevent reconnection if no auth token.
      onOpen: () => {
        const authToken = getAuthToken();
        if (authToken) {
          sendJsonMessage({
            type: 'authentication',
            payload: { access_token: authToken.access_token },
          });
        }
      },
    });

  return (
    <WebSocketContext.Provider
      value={{ sendJsonMessage, lastJsonMessage, readyState }}
    >
      {props.children}
    </WebSocketContext.Provider>
  );
}
