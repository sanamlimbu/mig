import { getPrivateMessages } from '@/api/user';
import { useAuth } from '@/hooks/auth';
import { PrivateMessage, User } from '@/types';
import { PersonIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import PrivateChat from './privateChat';
import { Avatar, AvatarImage } from './ui/avatar';
import { Input } from './ui/input';
import { ScrollArea } from './ui/scroll-area';

export default function PrivateChats() {
  const { user } = useAuth();
  const [recipient, setRecipient] = useState<User>();

  const { isPending, isError, data, error } = useQuery({
    queryKey: ['private-messages'],
    queryFn: () => {
      if (user === null) {
        return undefined;
      }
      return getPrivateMessages(user.id, { page: 1, page_size: 40 });
    },
  });

  const handleSearchTermChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    console.log(e);
  };

  if (isPending) {
    return <div>Loading</div>;
  }

  if (isError) {
    return <div>{error.message};</div>;
  }

  if (user === null) {
    return <div>{'error'}</div>;
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
        <ScrollArea className="flex-grow">
          <div>
            {data?.data.map((msg) => (
              <div
                key={msg.id}
                className="cursor-pointer hover:bg-slate-100 w-full"
              >
                <PrivateChatItem
                  user={user}
                  message={msg}
                  updateRecipient={(recipient) => setRecipient(recipient)}
                />
              </div>
            ))}
          </div>
        </ScrollArea>
      </div>
      <div className="w-full flex-grow min-w-96">
        {recipient && <PrivateChat user={user} recipient={recipient} />}
      </div>
    </div>
  );
}

interface PrivateChatItemProps {
  user: User;
  message: PrivateMessage;
  updateRecipient: (recipient: User) => void;
}

function PrivateChatItem({
  user,
  message,
  updateRecipient,
}: PrivateChatItemProps) {
  const recipient: User =
    user.username === message.recipient_username
      ? {
          id: message.sender_id,
          username: message.sender_username,
          email: message.sender_email,
          workflow_state: message.sender_workflow_state,
          avatar_url: '',
        }
      : {
          id: message.recipient_id,
          username: message.recipient_username,
          email: message.recipient_email,
          workflow_state: message.recipient_workflow_state,
          avatar_url: '',
        };

  const convetDateToFormattedString = (str: string) => {
    const date = new Date(str);

    const year = date.getFullYear();
    const month = date.getMonth() + 1;
    const day = date.getDate();

    return `${day}/${month}/${year}`;
  };

  return (
    <div
      onClick={() => updateRecipient(recipient)}
      className="text-gray-800 p-4 max-w-md"
    >
      <div className="flex items-center gap-4">
        <Avatar>
          <div className="rounded-full w-10 h-10 flex-shrink-0 bg-red-200 flex items-center justify-center">
            {recipient.avatar_url ? (
              <AvatarImage src={recipient.avatar_url} />
            ) : (
              <PersonIcon className="w-7 h-7" />
            )}
          </div>
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex justify-between">
            <p className="font-bold text-sm">{recipient.username}</p>
            <p className="text-xs">
              {convetDateToFormattedString(message.created_at)}
            </p>
          </div>
          <p className="text-sm truncate">{message.content}</p>
        </div>
      </div>
    </div>
  );
}
