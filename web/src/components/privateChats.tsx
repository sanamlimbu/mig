import { getRecentPrivateMessages } from '@/api/user';
import { useAuth } from '@/hooks/auth';
import { Message, User } from '@/types';
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
    queryKey: [user.id, 'recent-private-messages'],
    queryFn: () =>
      getRecentPrivateMessages(user.id, { page: 1, page_size: 40 }),
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
            {data?.map((msg) => (
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
  message: Message;
  updateRecipient: (recipient: User | undefined) => void;
}

function PrivateChatItem({
  user,
  message,
  updateRecipient,
}: PrivateChatItemProps) {
  const recipient =
    user.id === message.sender_id ? message.recipient : message.sender;

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
              {convetDateToFormattedString(message.created_at)}
            </p>
          </div>
          <p className="text-sm truncate">{message.content}</p>
        </div>
      </div>
    </div>
  );
}
