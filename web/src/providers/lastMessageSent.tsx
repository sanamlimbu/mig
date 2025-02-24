import { LastMessageSentContext } from '@/contexts/lastMessageSent';
import { Message } from '@/types';
import { PropsWithChildren, useState } from 'react';

export const LastMessageSentProvider = (props: PropsWithChildren) => {
  const [message, setMessage] = useState<Partial<Message>>();
  return (
    <LastMessageSentContext.Provider
      value={{ lastMessageSent: message, setLastMessageSent: setMessage }}
    >
      {props.children}
    </LastMessageSentContext.Provider>
  );
};
