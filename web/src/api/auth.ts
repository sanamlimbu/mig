import { User } from '@/types';
import { getAuthToken } from '@/utils/auth';
import { axios } from '../axios';

export interface AuthToken {
  access_token: string;
  refresh_token: string;
  user: User;
}

export async function login(username: string, password: string) {
  const resp = await axios.post<AuthToken>('/auth/login', {
    username: username,
    password: password,
  });
  return resp.data;
}

export async function refreshAccessToken(): Promise<AuthToken> {
  const authToken = getAuthToken();

  if (!authToken) {
    throw new Error('Missing authentication token.');
  }

  const response = await axios.post<AuthToken>('/auth/refresh-token', {
    access_token: authToken.access_token,
    refresh_token: authToken.refresh_token,
  });

  return response.data;
}
