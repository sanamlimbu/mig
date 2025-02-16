export interface Pagination {
  page: number;
  page_size: number;
}

type UserWorkflowState = 'active' | 'suspended' | 'unverified' | 'deleted';

type UserRole = 'superadmin' | 'admin' | 'member';
export interface User {
  id: string;
  username: string;
  email: string;
  workflow_state: UserWorkflowState;
  avatar_url: string;
  role: UserRole;
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}

export type MessageWorkflowState = 'created' | 'updated' | 'deleted';

export type MessageType = 'private' | 'chatroom';
export interface Message {
  id: string;
  workflow_state: MessageWorkflowState;
  is_read: boolean | null;
  content: string;
  recipient_id: string | null;
  chatroom_id: string | null;
  sender_id: string;
  sender?: User;
  chatroom?: Chatroom;
  recipient?: User;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

type ChatroomWorkflowState = 'active' | 'deleted';

type ChatroomType = 'private' | 'public';
export interface Chatroom {
  id: string;
  name: string;
  workflow_state: ChatroomWorkflowState;
  type: ChatroomType;
  avatar_url: string;
  created_by: string;
  creator?: User;
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}

interface AuthenticationPayload {
  access_token: string;
}
interface MessageCreatedPayload {
  id: string;
  sender_id: string;
  recipient_id: string;
  content: string;
  message_type: MessageType;
}

type WebSocketMessageType =
  | 'authentication'
  | 'message_created'
  | 'message_updated'
  | 'message_deleted';

type WebSocketMessagePayload = AuthenticationPayload | MessageCreatedPayload;

export interface WebSocketMessage {
  type: WebSocketMessageType;
  payload: WebSocketMessagePayload;
}
