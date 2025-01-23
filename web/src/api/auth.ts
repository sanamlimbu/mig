import { User } from '@/types';
import { getAuthToken } from '@/utils/auth';
import { axios } from '../axios';

export interface AuthToken {
  access_token: string;
  expires_at: number;
  expires_in: string;
  refresh_token: string;
  user: User;
}

export function login(username: string, password: string) {
  return axios.post<AuthToken>('/login', {
    username: username,
    password: password,
  });
}

interface RefreshAccessTokenResponse {
  access_token: string;
}

export function refreshAccessToken() {
  const authToken = getAuthToken();

  if (!authToken) {
    throw new Error('Missing authentication token.');
  }

  return axios.post<RefreshAccessTokenResponse>('/refresh-token', {
    access_token: authToken.access_token,
    refresh_token: authToken.refresh_token,
  });
}
