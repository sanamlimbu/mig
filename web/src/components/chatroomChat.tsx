import { getChatroomMessages } from '@/api/chatroom';
import { Avatar, AvatarImage } from '@/components/ui/avatar';
import { ScrollArea } from '@/components/ui/scroll-area';
import { WS_BASE_URL } from '@/constants';
import {
  Chatroom,
  Message,
  MessageCreatedPayload,
  User,
  WebSocketMessage,
} from '@/types';
import { getAuthToken } from '@/utils/auth';
import { convertDateToFormattedString } from '@/utils/helpers';
import { DotsVerticalIcon, PersonIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import useWebSocket from 'react-use-websocket';
import { v4 as uuidv4 } from 'uuid';
import SendIcon from '../assets/send.svg';
import { AlertError } from './ui/alert-error';
import { CenterDiv } from './ui/center-div';
import { LoadingSpinner } from './ui/loading-spinner';
import { Textarea } from './ui/textarea';

interface ChatroomProps {
  user: User;
  chatroom: Chatroom;
}
export default function ChatroomChat({ user, chatroom }: ChatroomProps) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: ['chatrooms', chatroom.id, 'messages'],
    queryFn: () =>
      getChatroomMessages(chatroom.id, {
        page: 1,
        page_size: 20,
      }),
  });

  if (isPending) {
    return (
      <CenterDiv>
        <LoadingSpinner />
      </CenterDiv>
    );
  }

  if (isError) {
    return <AlertError title="Error" message={error.message} />;
  }

  return <ChatBox user={user} chatroom={chatroom} recentMessages={data} />;
}

function ChatBox({
  user,
  chatroom,
  recentMessages,
}: {
  user: User;
  chatroom: Chatroom;
  recentMessages: Message[];
}) {
  const [messages, setMessages] = useState<Partial<Message>[]>([]);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const { sendJsonMessage, lastJsonMessage } = useWebSocket<WebSocketMessage>(
    WS_BASE_URL,
    {
      share: true,
      shouldReconnect: () => !!getAuthToken(), // Prevent reconnection if no auth token.
    }
  );
  useEffect(() => setMessages(recentMessages), [recentMessages]);
  useEffect(() => {
    if (lastJsonMessage && lastJsonMessage.type === 'message_created') {
      const payload = lastJsonMessage.payload as MessageCreatedPayload;
      if (payload.type === 'chatroom' && payload.recipient_id === chatroom.id) {
        const message: Partial<Message> = {
          id: payload.id,
          content: payload.content,
          recipient_id: payload.recipient_id,
          sender_id: payload.sender_id,
          type: payload.type,
          created_at: payload.created_at,
        };
        setMessages((prev) => [message, ...prev]);
      }
    }
  }, [chatroom.id, lastJsonMessage]);

  const handleSend = () => {
    if (!inputRef.current) {
      return;
    }

    const message: MessageCreatedPayload = {
      id: uuidv4(),
      sender_id: user.id,
      recipient_id: chatroom.id,
      content: inputRef.current?.value,
      type: 'chatroom',
    };

    sendJsonMessage<WebSocketMessage>(
      {
        type: 'message_created',
        payload: message,
      },
      true
    );

    setMessages((prev) => [message, ...prev]);
  };
  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between shadow-sm p-4 bg-gray-100">
        <div className="flex items-center gap-2 cursor-pointer">
          <Avatar>
            <div className="rounded-full w-10 h-10 flex-shrink-0 bg-red-200 flex items-center justify-center">
              {chatroom.avatar_url ? (
                <AvatarImage src={chatroom.avatar_url} />
              ) : (
                <PersonIcon className="w-7 h-7" />
              )}
            </div>
          </Avatar>
          <span className="font-medium">{chatroom.name}</span>
        </div>
        <div className="font-bold cursor-pointer">
          <DotsVerticalIcon className="w-5 h-5" />
        </div>
      </div>
      <ScrollArea className="pr-2 bg-slate-50 flex-grow">
        <div className="px-3 pt-3 flex flex-col-reverse">
          {messages?.map((msg) => {
            const isSentByUser = msg.sender_id === user.id;
            return (
              <div
                key={msg.id}
                className="flex items-center gap-2 mb-2.5 text-sm justify-between"
              >
                <div
                  className={`px-3 py-1 rounded-lg ${
                    isSentByUser
                      ? 'bg-cyan-600 text-white'
                      : 'bg-white text-gray-800'
                  }`}
                >
                  <p>{msg.content}</p>
                </div>
                <div className="text-xs">
                  <p className="bg-slate-700 text-white px-1 rounded-sm inline-block truncate max-w-[100px]">
                    {msg.sender?.username}
                  </p>
                  <p className="whitespace-nowrap">
                    {msg.created_at &&
                      convertDateToFormattedString(msg.created_at, true)}
                  </p>
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
