import { LastMessageSentContext } from '@/contexts/lastMessageSent';
import { useContext } from 'react';

export const useLastMessageSent = () => useContext(LastMessageSentContext);
