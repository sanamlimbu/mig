import type { WebSocketMessage } from './types';

export async function handleIncomingWebsocketMessage(
  message: WebSocketMessage
) {
  if (!message) {
    return;
  }

  const { type } = message;

  switch (type) {
    case 'authentication': {
      break;
    }
    case 'message_created': {
      break;
    }
    case 'message_updated': {
      break;
    }
    case 'message_deleted': {
      break;
    }
    default: {
      throw new Error(`Invalid WebSocket messsage type: ${type}`);
    }
  }
}
