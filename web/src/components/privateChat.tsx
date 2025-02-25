import { getPrivateConversation, updateReadMessages } from '@/api/user';
import { Avatar, AvatarImage } from '@/components/ui/avatar';
import { ScrollArea } from '@/components/ui/scroll-area';
import { WS_BASE_URL } from '@/constants';
import { useLastMessageSent } from '@/hooks/lastMessageSent';
import { queryClient } from '@/main';
import {
  Message,
  MessageCreatedPayload,
  User,
  WebSocketMessage,
} from '@/types';
import { getAuthToken } from '@/utils/auth';
import { DotsVerticalIcon, PersonIcon } from '@radix-ui/react-icons';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import useWebSocket from 'react-use-websocket';
import { v4 as uuidv4 } from 'uuid';
import SendIcon from '../assets/send.svg';
import { AlertError } from './ui/alert-error';
import { CenterDiv } from './ui/center-div';
import { LoadingSpinner } from './ui/loading-spinner';
import { Textarea } from './ui/textarea';

// https://github.com/radix-ui/primitives/discussions/990

interface PrivateChatProps {
  sender: User;
  recipient: User;
}

export default function PrivateChat({ sender, recipient }: PrivateChatProps) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: [sender.id, recipient.id, 'private-conversation'],
    queryFn: () =>
      getPrivateConversation(sender.id, recipient.id, {
        page: 1,
        page_size: 20,
      }),
  });
  const isUpdateReadMessagesCompletedRef = useRef(false);
  const { mutate } = useMutation({
    mutationFn: () => updateReadMessages(recipient.id, sender.id),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [sender.id, recipient.id, 'private-conversation'],
      });
      isUpdateReadMessagesCompletedRef.current = true;
    },
  });

  useEffect(() => {
    if (!data || isUpdateReadMessagesCompletedRef.current) {
      return;
    }

    const hasSomeUnreadMessages = data.some(
      (m) => m.sender_id === recipient.id && m.is_read === false
    );

    if (hasSomeUnreadMessages) {
      mutate();
    }
  }, [data, mutate, recipient.id]);

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

  return (
    <ChatBox sender={sender} recipient={recipient} recentMessages={data} />
  );
}

function ChatBox({
  sender,
  recipient,
  recentMessages,
}: {
  sender: User;
  recipient: User;
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
  const { setLastMessageSent } = useLastMessageSent();

  useEffect(() => setMessages(recentMessages), [recentMessages]);

  useEffect(() => {
    if (lastJsonMessage && lastJsonMessage.type === 'message_created') {
      const payload = lastJsonMessage.payload as MessageCreatedPayload;
      if (payload.type === 'private' && payload.sender_id === recipient.id) {
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
  }, [lastJsonMessage, recipient.id]);

  const handleSend = () => {
    if (!inputRef.current) {
      return;
    }

    const message: MessageCreatedPayload = {
      id: uuidv4(),
      sender_id: sender.id,
      recipient_id: recipient.id,
      content: inputRef.current?.value,
      type: 'private',
      sender_username: sender.username,
      recipient_name: recipient.username,
    };

    sendJsonMessage<WebSocketMessage>(
      {
        type: 'message_created',
        payload: message,
      },
      true
    );

    setMessages((prev) => [message, ...prev]);
    setLastMessageSent({ ...message, created_at: new Date().toISOString() });
  };

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
          {messages?.map((msg) => {
            const isSentByUser = msg.sender_id === sender.id;
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
