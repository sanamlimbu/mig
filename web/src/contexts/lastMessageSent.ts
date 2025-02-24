import { Message } from '@/types';
import { createContext } from 'react';

interface ILastMessageSentContext {
  lastMessageSent: Partial<Message> | undefined;
  setLastMessageSent: React.Dispatch<
    React.SetStateAction<Partial<Message> | undefined>
  >;
}

const defaultContextValue: ILastMessageSentContext = {
  lastMessageSent: {},
  setLastMessageSent: () => {},
};

export const LastMessageSentContext =
  createContext<ILastMessageSentContext>(defaultContextValue);
