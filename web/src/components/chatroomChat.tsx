import { getChatroomMessages } from '@/api/chatroom';
import { Avatar, AvatarImage } from '@/components/ui/avatar';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Chatroom, User } from '@/types';
import { DotsVerticalIcon, PersonIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useRef, useState } from 'react';
import SendIcon from '../assets/send.svg';
import { Textarea } from './ui/textarea';

interface ChatroomProps {
  user: User;
  chatroom: Chatroom;
}
export default function ChatroomChat({ user, chatroom }: ChatroomProps) {
  const [content, setContent] = useState('');
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const { isPending, isError, data, error } = useQuery({
    queryKey: ['chatrooms', chatroom.id, 'messages'],
    queryFn: () =>
      getChatroomMessages(chatroom.id, {
        page: 1,
        page_size: 20,
      }),
  });

  const sendMessage = () => {
    console.log(content);
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
          {data?.data.map((msg) => {
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
        <Textarea
          className="border-white bg-white"
          onChange={(e) => setContent(e.currentTarget.value)}
          ref={inputRef}
        />
        <img
          src={SendIcon}
          className="w-7 h-7 cursor-pointer"
          onClick={() => {
            if (inputRef.current) {
              inputRef.current.value = '';
            }
            sendMessage();
          }}
        />
      </div>
    </div>
  );
}
