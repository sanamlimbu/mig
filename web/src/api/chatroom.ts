import { axios } from '@/axios';
import { Chatroom, Message, Pagination } from '@/types';

export async function getChatrooms(searchTerm: string, pagination: Pagination) {
  const { page, page_size } = pagination;
  const resp = await axios.get<Chatroom[]>(
    searchTerm
      ? `/chatrooms?page=${page}&page_size=${page_size}&search_term=${searchTerm}`
      : `/chatrooms?page=${page}&page_size=${page_size}`
  );

  return resp.data;
}

export async function getChatroomMessages(
  chatroomID: string,
  pagination: Pagination
) {
  const { page, page_size } = pagination;
  const resp = await axios.get<Message[]>(
    `/chatrooms/${chatroomID}/messages?page=${page}&page_size=${page_size}`
  );
  return resp.data;
}
