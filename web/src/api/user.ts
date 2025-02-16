import { axios } from '@/axios';
import { Message, Pagination } from '@/types';

export function getPrivateConversationQueryKey(
  userID: string,
  recipientID: string
) {
  return [userID, 'private-conversation', recipientID];
}

export async function getRecentPrivateMessages(
  userID: string,
  pagination: Pagination
) {
  const { page, page_size } = pagination;
  const resp = await axios.get<Message[]>(
    `/users/${userID}/recent-private-messages?page=${page}&page_size=${page_size}`
  );
  return resp.data;
}

export async function getPrivateConversation(
  userID: string,
  recipientID: string,
  pagination: Pagination
) {
  const { page, page_size } = pagination;
  const resp = await axios.get<Message[]>(
    `/users/${userID}/private-conversation/${recipientID}?page=${page}&page_size=${page_size}`
  );
  return resp.data;
}
