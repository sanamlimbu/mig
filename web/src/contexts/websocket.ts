import { WebSocketMessage } from '@/types';
import { createContext, useContext } from 'react';
import { ReadyState } from 'react-use-websocket';
import { SendJsonMessage } from 'react-use-websocket/dist/lib/types';

interface WebSocketContextValue {
  sendJsonMessage: SendJsonMessage;
  lastJsonMessage: WebSocketMessage;
  readyState: ReadyState;
}

export const WebSocketContext = createContext<WebSocketContextValue | null>(
  null
);

export const useWebSocketContext = () => useContext(WebSocketContext);
