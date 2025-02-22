import { getRecentPrivateMessages } from '@/api/user';
import { WS_BASE_URL } from '@/constants';
import { useAuth } from '@/hooks/auth';
import {
  Message,
  MessageCreatedPayload,
  User,
  WebSocketMessage,
} from '@/types';
import { convertDateToFormattedString } from '@/utils/helpers';
import { PersonIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
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
  const [selectedRecipient, setSelectedRecipient] = useState<User>();

  const { isPending, isError, data, error } = useQuery({
    queryKey: [user.id, 'recent-private-messages'],
    queryFn: () =>
      getRecentPrivateMessages(user.id, { page: 1, page_size: 40 }),
  });

  const handleSearchTermChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    console.log(e);
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

  return (
    <div className="flex w-full overflow-x-auto">
      <div className="flex flex-col">
        <div className="px-4 mt-4 max-w-md">
          <p className="text-xl font-bold">Chats</p>
          <Input
            type="text"
            className="mt-3 mb-2"
            onChange={handleSearchTermChange}
            placeholder="Search"
          />
        </div>
        <ScrollArea className="h-[100vh] flex-grow">
          <div>
            {data?.map((msg) => {
              const recipient =
                user.id === msg.sender_id ? msg.recipient : msg.sender;
              return (
                <div
                  key={msg.id}
                  className={`cursor-pointer hover:bg-slate-100 w-full ${
                    selectedRecipient?.id === recipient?.id && 'bg-slate-100'
                  }`}
                >
                  <PrivateChatItem
                    recipient={recipient!}
                    message={msg}
                    selectedRecipient={selectedRecipient}
                    updateRecipientSelection={(recipient) =>
                      setSelectedRecipient(recipient)
                    }
                  />
                </div>
              );
            })}
          </div>
        </ScrollArea>
      </div>
      <div className="w-full flex-grow min-w-96">
        {selectedRecipient && (
          <PrivateChat user={user} recipient={selectedRecipient} />
        )}
      </div>
    </div>
  );
}

interface PrivateChatItemProps {
  recipient: User;
  message: Message;
  selectedRecipient: User | undefined;
  updateRecipientSelection: (recipient: User | undefined) => void;
}

function PrivateChatItem({
  recipient,
  message,
  selectedRecipient,
  updateRecipientSelection,
}: PrivateChatItemProps) {
  const [messages, setMessages] = useState<Partial<Message>[]>([message]);
  const { lastJsonMessage } = useWebSocket<WebSocketMessage>(WS_BASE_URL, {
    share: true,
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
      };
      console.log(message);

      if (message.sender_id === recipient.id) {
        if (message.sender_id === selectedRecipient?.id) {
          setMessages((prev) => [{ ...message, is_read: true }, ...prev]);
        } else {
          setMessages((prev) => [{ ...message, is_read: false }, ...prev]);
        }
      }
    }
  }, [lastJsonMessage, recipient.id, selectedRecipient?.id]);

  const unreadMessages = messages.filter((msg) => !msg.is_read);

  return (
    <div
      onClick={() => updateRecipientSelection(recipient)}
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
              {convertDateToFormattedString(message.created_at)}
            </p>
          </div>
          <div className="flex text-sm justify-between items-center gap-2">
            <p className="truncate flex-1">{message.content}</p>
            <p className="bg-green-500 text-white rounded-full min-w-[20px] h-5 flex items-center justify-center px-1 text-[10px]">
              {unreadMessages.length}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
