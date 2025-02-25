import { getChatrooms } from '@/api/chatroom';
import { useAuth } from '@/hooks/auth';
import { useDebounce } from '@/hooks/debounce';
import { Chatroom, User } from '@/types';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import ChatroomChat from './chatroomChat';
import { AlertError } from './ui/alert-error';
import { Avatar, AvatarImage } from './ui/avatar';
import { CenterDiv } from './ui/center-div';
import { Input } from './ui/input';
import { LoadingSpinner } from './ui/loading-spinner';
import { ScrollArea } from './ui/scroll-area';

export default function Chatrooms() {
  const { user } = useAuth();
  const [selectedChatroom, setSelectedChatroom] = useState<Chatroom>();
  const [searchTerm, setSearchTerm] = useState('');
  const debouncedSetSearchTerm = useDebounce(setSearchTerm);

  const { isPending, isError, data, error } = useQuery({
    queryKey: ['chatrooms', searchTerm],
    queryFn: () => getChatrooms(searchTerm, { page: 1, page_size: 20 }),
  });

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

  return (
    <div className="flex w-full overflow-x-auto">
      <div className="flex flex-col w-full max-w-md">
        <div className="px-4 mt-4">
          <p className="text-xl font-bold">Chatrooms</p>
          <Input
            type="text"
            className="mt-3 mb-2 w-full"
            onChange={handleSearchTermChange}
            placeholder="Search"
            defaultValue={searchTerm}
          />
        </div>
        <ScrollArea className="flex-grow">
          <div>
            {data?.map((chatroom) => (
              <div
                key={chatroom.id}
                className={`cursor-pointer hover:bg-slate-100 w-full ${
                  selectedChatroom?.id === chatroom.id && 'bg-slate-100'
                }`}
                onClick={(e) => {
                  e.stopPropagation();
                  setSelectedChatroom(chatroom);
                }}
              >
                <ChatroomItem user={user} chatroom={chatroom} />
              </div>
            ))}
          </div>
        </ScrollArea>
      </div>
      <div className="w-full flex-grow min-w-96">
        {selectedChatroom && (
          <ChatroomChat user={user} chatroom={selectedChatroom} />
        )}
      </div>
    </div>
  );
}

interface ChatroomItemProps {
  user: User;
  chatroom: Chatroom;
}

function ChatroomItem({ chatroom }: ChatroomItemProps) {
  return (
    <div className="text-gray-800 p-4">
      <div className="flex items-center gap-4">
        <Avatar>
          <div className="rounded-full w-10 h-10 flex-shrink-0 bg-red-200 flex items-center justify-center">
            {chatroom.avatar_url ? (
              <AvatarImage src={chatroom.avatar_url} />
            ) : (
              <span>{chatroom.name[0].toUpperCase()}</span>
            )}
          </div>
        </Avatar>
        <div className="flex-1 min-w-0">
          <p className="font-bold text-sm my-1">{chatroom.name}</p>
        </div>
      </div>
    </div>
  );
}
