export interface Pagination {
  page: number;
  page_size: number;
}

type UserWorkflowState = 'active' | 'suspended' | 'unverified' | 'deleted';
export interface User {
  id: string;
  username: string;
  email: string;
  workflow_state: UserWorkflowState;
  avatar_url: string;
}

type MessageWorkflowState = 'created' | 'updated' | 'deleted';

export interface PrivateMessage {
  id: string;
  workflow_state: MessageWorkflowState;
  is_read: boolean;
  content: string;
  created_at: string;
  recipient_id: string;
  recipient_username: string;
  recipient_email: string;
  recipient_workflow_state: UserWorkflowState;
  sender_id: string;
  sender_email: string;
  sender_username: string;
  sender_workflow_state: UserWorkflowState;
}

type ChatroomWorkflowState = 'active' | 'deleted';

type ChatroomType = 'private' | 'public';
export interface Chatroom {
  id: string;
  name: string;
  workflow_state: ChatroomWorkflowState;
  type: ChatroomType;
  created_by: string;
  creator_username: string;
  creator_email: string;
  creator_workflow_state: UserWorkflowState;
}

export interface ChatroomMessage {
  id: string;
  workflow_state: MessageWorkflowState;
  content: string;
  created_at: string;
  chatroom_id: string;
  chatroon_workflow_state: ChatroomWorkflowState;
  chatroom_type: ChatroomType;
  sender_id: string;
  sender_email: string;
  sender_username: string;
  sender_workflow_state: UserWorkflowState;
}
