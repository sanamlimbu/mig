import { getChatrooms } from '@/api/chatroom';
import { useAuth } from '@/hooks/auth';
import { Chatroom, User } from '@/types';
import { PersonIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import ChatroomChat from './chatroomChat';
import { Avatar, AvatarImage } from './ui/avatar';
import { Input } from './ui/input';
import { ScrollArea } from './ui/scroll-area';
import { Separator } from './ui/separator';

export default function Chatrooms() {
  const { user } = useAuth();
  const [chatroom, setChatroom] = useState<Chatroom>();
  const [searchTerm, setSearchTerm] = useState('');

  const { isPending, isError, data, error } = useQuery({
    queryKey: ['chatrooms', searchTerm],
    queryFn: () => getChatrooms(searchTerm, { page: 1, page_size: 20 }),
  });

  if (isPending) {
    return <div>Loading</div>;
  }

  if (isError) {
    return <div>{error.message};</div>;
  }

  if (user === null) {
    return <div>{'error'}</div>;
  }

  console.log(data.data);

  return (
    <div className="flex w-full overflow-x-auto">
      <div className="flex flex-col w-full max-w-md">
        <div className="px-4 mt-4">
          <p className="text-xl font-bold">Chatrooms</p>
          <Input
            type="text"
            className="mt-3 mb-2 w-full"
            onChange={(e) => setSearchTerm(e.currentTarget.value)}
            placeholder="Search"
          />
        </div>
        <ScrollArea className="flex-grow">
          <div>
            {data?.data.map((chatroom) => (
              <div
                key={chatroom.id}
                className="cursor-pointer hover:bg-slate-100 w-full"
              >
                <ChatroomItem
                  user={user}
                  chatroom={chatroom}
                  updateChatroom={(chatroom) => setChatroom(chatroom)}
                />
              </div>
            ))}
          </div>
        </ScrollArea>
      </div>
      <div className="w-full flex-grow min-w-96">
        {chatroom && <ChatroomChat user={user} chatroom={chatroom} />}
      </div>
    </div>
  );
}

interface ChatroomItemProps {
  user: User;
  chatroom: Chatroom;
  updateChatroom: (chatroom: Chatroom) => void;
}

function ChatroomItem({ updateChatroom, chatroom }: ChatroomItemProps) {
  return (
    <div onClick={() => updateChatroom(chatroom)} className="text-gray-800 p-4">
      <div className="flex items-center gap-4">
        <Avatar>
          <div className="rounded-full w-10 h-10 flex-shrink-0 bg-red-200 flex items-center justify-center">
            {chatroom.avatar_url ? (
              <AvatarImage src={chatroom.avatar_url} />
            ) : (
              <PersonIcon className="w-7 h-7" />
            )}
          </div>
        </Avatar>
        <div className="flex-1 min-w-0">
          <Separator />
          <p className="font-bold text-sm my-1">{chatroom.name}</p>
          <Separator />
        </div>
      </div>
    </div>
  );
}
