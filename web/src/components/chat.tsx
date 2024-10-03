import { Button, TextArea } from '@radix-ui/themes';
import { FormEvent, useEffect, useState } from 'react';
import useWebSocket, { ReadyState } from 'react-use-websocket';
import { WS_BASE_URL } from '../constants';
import MessageBox from './message';

export default function Chat() {
  const [msg, setMsg] = useState('');
  const { readyState, sendJsonMessage, lastJsonMessage } = useWebSocket(
    WS_BASE_URL,
    {
      share: true,
      shouldReconnect: () => true,
    }
  );

  useEffect(() => {
    console.log(lastJsonMessage);
  }, [lastJsonMessage]);

  const connectionStatus = {
    [ReadyState.CONNECTING]: 'Connecting',
    [ReadyState.OPEN]: 'Open',
    [ReadyState.CLOSING]: 'Closing',
    [ReadyState.CLOSED]: 'Closed',
    [ReadyState.UNINSTANTIATED]: 'Uninstantiated',
  }[readyState];

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (msg && readyState === ReadyState.OPEN) {
      sendJsonMessage({
        sender_id: 1,
        receiver_id: 1,
        message_type: 'private',
        content: msg,
      });
    }
  };

  console.log(connectionStatus);
  return (
    <div className="m-4">
      <form onSubmit={handleSubmit} className="p-4">
        <TextArea
          placeholder="Enter message…"
          onChange={(e) => setMsg(e.currentTarget.value)}
          mb="2"
        />
        <Button type="submit">Submit</Button>
      </form>
      <MessageBox
        message={{
          id: 1,
          content: 'Hey meet me at the spot',
          sender_id: 1,
          sender_name: 'sudosanam',
          sender_workflow_state: 'active',
          receiver_id: 1,
          receiver_name: 'hello',
          receiver_workflow_state: 'active',
          created_at: new Date(),
          workflow_state: 'created',
        }}
      />
      <MessageBox
        message={{
          id: 1,
          content:
            'It is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. The point of using Lorem Ipsum is that',
          sender_id: 1,
          sender_name: 'sudosanam',
          sender_workflow_state: 'active',
          receiver_id: 1,
          receiver_name: 'hello',
          receiver_workflow_state: 'active',
          created_at: new Date(),
          workflow_state: 'created',
        }}
      />
    </div>
  );
}
