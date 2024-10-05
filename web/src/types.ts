export type UsersWorkflowState = 'active' | 'suspended' | 'deleted';

export interface User {
  id: number;
  username: string;
  worklow_state: UsersWorkflowState;
}

export type GroupsWorkflowState = 'active' | 'deleted';

export interface Group {
  id: number;
  name: string;
  workflow_state: GroupsWorkflowState;
}

type MessagesWorkflowState = 'created' | 'deleted';

export interface Message {
  id: number;
  content: string;
  workflow_state: MessagesWorkflowState;
  created_at: Date;
  sender_id: number;
  sender_name: string;
  sender_workflow_state: UsersWorkflowState;
  receiver_id: number;
  receiver_name: string;
  receiver_workflow_state: UsersWorkflowState | GroupsWorkflowState;
}
