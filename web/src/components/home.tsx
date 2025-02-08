import { WS_BASE_URL } from '@/constants';
import { useAuth } from '@/hooks/auth';
import { WebSocketMessage } from '@/types';
import { getAuthToken } from '@/utils/auth';
import {
  ChatBubbleIcon,
  Component1Icon,
  GroupIcon,
} from '@radix-ui/react-icons';
import { useState } from 'react';
import useWebSocket from 'react-use-websocket';
import Chatrooms from './chatrooms';
import Login from './login';
import PrivateChats from './privateChats';

type Menu = 'Chats' | 'Status' | 'Chatrooms';

export default function Home() {
  const { user } = useAuth();
  const [menu, setMenu] = useState<Menu>('Chats');
  const { sendJsonMessage } = useWebSocket<WebSocketMessage>(WS_BASE_URL, {
    share: true,
    onOpen: () => {
      const authToken = getAuthToken();
      if (!authToken) {
        return;
      }

      sendJsonMessage<WebSocketMessage>({
        type: 'authentication',
        payload: {
          access_token: authToken.access_token,
        },
      });
    },
    // Prevent reconnection if no auth token.
    shouldReconnect: () => !!getAuthToken(),
  });

  if (user === null) {
    return <Login />;
  }

  return (
    <div className="h-screen flex flex-col overflow-hidden relative">
      <div className="h-24 bg-cyan-500"></div>
      <div className="flex-grow bg-zinc-100"></div>
      <div className="absolute left-0 right-0 z-20 bg-white border m-5 flex flex-col h-[calc(100%-2.5rem)]">
        <div className="flex-grow flex overflow-hidden">
          <Menus menu={menu} updateMenu={(menu) => setMenu(menu)} />
          {menu === 'Chats' && <PrivateChats />}
          {menu === 'Status' && <Chatrooms />}
          {menu === 'Chatrooms' && <Chatrooms />}
        </div>
      </div>
    </div>
  );
}

interface MenusProps {
  menu: Menu;
  updateMenu: (menu: Menu) => void;
}
function Menus({ menu, updateMenu }: MenusProps) {
  const classes =
    'rounded-full flex items-center justify-center w-11 h-11 cursor-pointer ';

  return (
    <div className="flex flex-col gap-2 p-3 bg-gray-100">
      <div
        className={classes + (menu === 'Chats' ? 'bg-gray-300' : '')}
        onClick={() => updateMenu('Chats')}
      >
        <ChatBubbleIcon className="w-6 h-6" />
      </div>
      <div
        className={classes + (menu === 'Status' ? 'bg-gray-200' : '')}
        onClick={() => updateMenu('Status')}
      >
        <GroupIcon className="w-6 h-6" />
      </div>
      <div
        className={classes + (menu === 'Chatrooms' ? 'bg-gray-200' : '')}
        onClick={() => updateMenu('Chatrooms')}
      >
        <Component1Icon className="w-6 h-6" />
      </div>
    </div>
  );
}
