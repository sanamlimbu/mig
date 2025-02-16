import { Message } from '@/types';
import { createContext, PropsWithChildren, useState } from 'react';

interface RecentPrivateMessagesContext {
  messages: Partial<Message>[];
  setMessages: (messages: Partial<Message>[]) => void;
}

export const RecentPrivateMessagesContext = createContext<
  RecentPrivateMessagesContext | undefined
>(undefined);

export default function RecentPrivateMessagesProvider(
  props: PropsWithChildren
) {
  const [messages, setMessages] = useState<Partial<Message>[]>([]);
  return (
    <RecentPrivateMessagesContext.Provider
      value={{
        messages: messages,
        setMessages: (messages) => {
          setMessages((prev) => [...messages, ...prev]);
        },
      }}
    >
      {props.children}
    </RecentPrivateMessagesContext.Provider>
  );
}
