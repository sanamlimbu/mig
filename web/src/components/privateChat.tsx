import { getPrivateConversation } from '@/api/user';
import { Avatar, AvatarImage } from '@/components/ui/avatar';
import { ScrollArea } from '@/components/ui/scroll-area';
import { WS_BASE_URL } from '@/constants';
import { User, WebSocketMessage } from '@/types';
import { getAuthToken } from '@/utils/auth';
import { DotsVerticalIcon, PersonIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useEffect, useRef } from 'react';
import useWebSocket from 'react-use-websocket';
import { v4 as uuidv4 } from 'uuid';
import SendIcon from '../assets/send.svg';
import { Textarea } from './ui/textarea';

interface PrivateChatProps {
  user: User;
  recipient: User;
}
export default function PrivateChat({ user, recipient }: PrivateChatProps) {
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const { sendJsonMessage, lastJsonMessage } = useWebSocket<WebSocketMessage>(
    WS_BASE_URL,
    {
      share: true,
      shouldReconnect: () => !!getAuthToken(),
    }
  );

  useEffect(() => {}, [lastJsonMessage]);

  const { isPending, isError, data, error } = useQuery({
    queryKey: [user.id, 'private-conversation', recipient.id],
    queryFn: () => {
      if (user === null) {
        return undefined;
      }
      return getPrivateConversation(user.id, recipient.id, {
        page: 1,
        page_size: 20,
      });
    },
  });

  const handleSend = () => {
    if (!inputRef.current) {
      return;
    }

    sendJsonMessage<WebSocketMessage>({
      type: 'message_created',
      payload: {
        id: uuidv4(),
        sender_id: user.id,
        recipient_id: recipient.id,
        content: inputRef.current?.value,
        message_type: 'private',
      },
    });
  };

  if (isPending) {
    return <div>Loading</div>;
  }

  if (isError) {
    return <div>{error.message};</div>;
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between shadow-sm p-4 bg-gray-100">
        <div className="flex items-center gap-2 cursor-pointer">
          <Avatar>
            <div className="rounded-full w-10 h-10 flex-shrink-0 bg-red-200 flex items-center justify-center">
              {recipient.avatar_url ? (
                <AvatarImage src={recipient.avatar_url} />
              ) : (
                <PersonIcon className="w-7 h-7" />
              )}
            </div>
          </Avatar>
          <span className="font-medium">{recipient.username}</span>
        </div>
        <div className="font-bold cursor-pointer">
          <DotsVerticalIcon className="w-5 h-5" />
        </div>
      </div>
      <ScrollArea className="pr-2 bg-slate-50 flex-grow">
        <div className="px-3 pt-3 flex flex-col-reverse">
          {data?.map((msg) => {
            const isSentByUser = msg.sender_username === user.username;
            return (
              <div
                key={msg.id}
                className={`flex mb-4 ${
                  isSentByUser ? 'justify-end' : 'justify-start'
                }`}
              >
                <div
                  className={`max-w-sm px-4 py-2 rounded-lg ${
                    isSentByUser
                      ? 'bg-cyan-600 text-white'
                      : 'bg-white text-gray-800'
                  }`}
                >
                  <p className="text-sm">{msg.content}</p>
                </div>
              </div>
            );
          })}
        </div>
      </ScrollArea>
      <div className="py-4 pl-4 pr-1 bg-gray-100 flex justify-between gap-3 items-center">
        <Textarea className="border-white bg-white" ref={inputRef} />
        <img
          src={SendIcon}
          className="w-7 h-7 cursor-pointer"
          onClick={() => {
            handleSend();
            if (inputRef.current) {
              inputRef.current.value = '';
            }
          }}
        />
      </div>
    </div>
  );
}
