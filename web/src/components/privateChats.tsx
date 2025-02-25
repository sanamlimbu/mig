import { getRecentPrivateMessages } from '@/api/user';
import { WS_BASE_URL } from '@/constants';
import { useAuth } from '@/hooks/auth';
import { useDebounce } from '@/hooks/debounce';
import { useLastMessageSent } from '@/hooks/lastMessageSent';
import { LastMessageSentProvider } from '@/providers/lastMessageSent';
import {
  Message,
  MessageCreatedPayload,
  User,
  WebSocketMessage,
} from '@/types';
import { getAuthToken } from '@/utils/auth';
import { convertDateToFormattedString } from '@/utils/helpers';
import { PersonIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import useWebSocket from 'react-use-websocket';
import PrivateChat from './privateChat';
import { AlertError } from './ui/alert-error';
import { Avatar, AvatarImage } from './ui/avatar';
import { CenterDiv } from './ui/center-div';
import { Input } from './ui/input';
import { LoadingSpinner } from './ui/loading-spinner';
import { ScrollArea } from './ui/scroll-area';

export default function PrivateChats() {
  const { user } = useAuth();
  const [currentRecipient, setCurrentRecipient] = useState<User>();
  const [searchTerm, setSearchTerm] = useState('');
  const debouncedSetSearchTerm = useDebounce(setSearchTerm);

  const { isPending, isError, data, error } = useQuery({
    queryKey: [user.id, searchTerm, 'recent-private-messages'],
    queryFn: () =>
      getRecentPrivateMessages(user.id, searchTerm, { page: 1, page_size: 40 }),
  });
  const currentRecipientRef = useRef(currentRecipient);
  currentRecipientRef.current = currentRecipient;

  const handleSearchTermChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    debouncedSetSearchTerm(e.currentTarget.value);
  };

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

  const getRecipientFromMessage = (currentUser: User, message: Message) => {
    if (currentUser.id === message.sender_id) {
      return message.recipient;
    }
    return message.sender;
  };

  const removeDuplicateMessages = (messages: Message[]) => {
    const result: Message[] = [];
    const seen = new Set<string>();

    for (let i = 0; i < messages.length; i++) {
      if (!seen.has(messages[i].id)) {
        seen.add(messages[i].id);
        result.push(messages[i]);
      }
    }

    return result;
  };

  return (
    <LastMessageSentProvider>
      <div className="flex w-full overflow-x-auto">
        <div className="flex flex-col">
          <div className="px-4 mt-4 max-w-md">
            <p className="text-xl font-bold">Chats</p>
            <Input
              type="text"
              className="mt-3 mb-2"
              onChange={handleSearchTermChange}
              placeholder="Search"
              defaultValue={searchTerm}
            />
          </div>
          <ScrollArea className="h-[100vh] flex-grow">
            <div>
              {data?.map((d) => {
                const recipient = getRecipientFromMessage(user, d.message);
                const messages =
                  d.unread_messages.length === 0
                    ? [d.message]
                    : removeDuplicateMessages([
                        d.message,
                        ...d.unread_messages,
                      ]);
                return (
                  <div
                    key={d.message.id}
                    className={`cursor-pointer hover:bg-slate-100 w-full ${
                      currentRecipient?.id === recipient?.id && 'bg-slate-100'
                    }`}
                    onClick={(e) => {
                      e.stopPropagation();
                      setCurrentRecipient(recipient);
                    }}
                  >
                    <PrivateChatItem
                      recipient={recipient!}
                      recentMessages={messages}
                      currentRecipientRef={currentRecipientRef}
                    />
                  </div>
                );
              })}
            </div>
          </ScrollArea>
        </div>
        <div className="w-full flex-grow min-w-96">
          {currentRecipient && (
            <PrivateChat sender={user} recipient={currentRecipient} />
          )}
        </div>
      </div>
    </LastMessageSentProvider>
  );
}

interface PrivateChatItemProps {
  recipient: User;
  recentMessages: Message[];
  currentRecipientRef: React.MutableRefObject<User | undefined>;
}

function PrivateChatItem({
  recipient,
  recentMessages,
  currentRecipientRef,
}: PrivateChatItemProps) {
  const { user } = useAuth();
  const { lastMessageSent } = useLastMessageSent();
  const [messages, setMessages] = useState<Partial<Message>[]>(recentMessages);
  const { lastJsonMessage } = useWebSocket<WebSocketMessage>(WS_BASE_URL, {
    share: true,
    shouldReconnect: () => !!getAuthToken(), // Prevent reconnection if no auth token.
  });

  useEffect(() => {
    if (lastJsonMessage && lastJsonMessage.type === 'message_created') {
      const payload = lastJsonMessage.payload as MessageCreatedPayload;
      const message: Partial<Message> = {
        id: payload.id,
        content: payload.content,
        recipient_id: payload.recipient_id,
        sender_id: payload.sender_id,
        type: payload.type,
        created_at: payload.created_at,
      };

      if (message.sender_id === recipient.id) {
        if (message.sender_id === currentRecipientRef.current?.id) {
          setMessages((prev) => [{ ...message, is_read: true }, ...prev]);
        } else {
          setMessages((prev) => [{ ...message, is_read: false }, ...prev]);
        }
      }
    }
  }, [currentRecipientRef, lastJsonMessage, recipient.id]);

  const unreadMessagesCount = messages.reduce((acc, msg) => {
    if (msg.recipient_id === user.id && msg.is_read === false) {
      return acc + 1;
    }
    return acc;
  }, 0);

  const getFirstMessage = (
    recipient: User,
    lastMessageSent: Partial<Message> | undefined,
    recentMessages: Partial<Message>[]
  ) => {
    if (lastMessageSent?.recipient_id !== recipient.id) {
      return recentMessages[0];
    }

    const sortedMessages = [lastMessageSent, ...recentMessages].sort(
      (a, b) =>
        new Date(b.created_at!).getTime() - new Date(a.created_at!).getTime()
    );

    return sortedMessages[0];
  };

  const message = getFirstMessage(recipient, lastMessageSent, messages);

  return (
    <div
      onClick={() => {
        const readMessages = messages.map((m) => ({
          ...m,
          is_read: true,
        }));
        setMessages(readMessages);
      }}
      className="text-gray-800 p-4 max-w-md"
    >
      <div className="flex items-center gap-4">
        <Avatar>
          <div className="rounded-full w-10 h-10 flex-shrink-0 bg-red-200 flex items-center justify-center">
            {recipient?.avatar_url ? (
              <AvatarImage src={recipient.avatar_url} />
            ) : (
              <PersonIcon className="w-7 h-7" />
            )}
          </div>
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex justify-between">
            <p className="font-bold text-sm">{recipient?.username}</p>
            <p className="text-xs">
              {message.created_at &&
                convertDateToFormattedString(message.created_at)}
            </p>
          </div>
          <div className="flex text-sm justify-between items-center gap-2">
            <p className="truncate flex-1">{message.content}</p>
            {unreadMessagesCount > 0 && (
              <p className="bg-green-500 text-white rounded-full min-w-[20px] h-5 flex items-center justify-center px-1 text-[10px]">
                {unreadMessagesCount}
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
