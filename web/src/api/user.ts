import { axios } from '@/axios';
import { Pagination, PrivateMessage } from '@/types';

export function getPrivateMessages(userID: string, pagination: Pagination) {
  const { page, page_size } = pagination;
  return axios.get<PrivateMessage[]>(
    `/users/${userID}/private-messages?page=${page}&page_size=${page_size}`
  );
}

export function getPrivateConversation(
  userID: string,
  recipientID: string,
  pagination: Pagination
) {
  const { page, page_size } = pagination;
  return axios.get<PrivateMessage[]>(
    `/users/${userID}/private-conversation/${recipientID}?page=${page}&page_size=${page_size}`
  );
}
