import { axios } from '@/axios';
import { Chatroom, ChatroomMessage, Pagination } from '@/types';

export function getChatrooms(searchTerm: string, pagination: Pagination) {
  const { page, page_size } = pagination;
  return axios.get<Chatroom[]>(
    searchTerm
      ? `/chatrooms?page=${page}&page_size=${page_size}&search_term=${searchTerm}`
      : `/chatrooms?page=${page}&page_size=${page_size}`
  );
}

export function getChatroomMessages(
  chatroomID: string,
  pagination: Pagination
) {
  const { page, page_size } = pagination;
  return axios.get<ChatroomMessage[]>(
    `/chatrooms/${chatroomID}/messages?page=${page}&page_size=${page_size}`
  );
}
