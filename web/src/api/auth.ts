import { User } from '@/types';
import { axios } from '../axios';

interface LoginResponse {
  access_token: string;
  expires_at: number;
  expires_in: string;
  refresh_token: string;
  user: User;
}

export function login(username: string, password: string) {
  return axios.post<LoginResponse>('/login', {
    username: username,
    password: password,
  });
}

interface RefreshAccessTokenResponse {
  access_token: string;
}

export function refreshAccessToken() {
  const accessToken = window.sessionStorage.getItem('access-token');
  const refreshToken = window.sessionStorage.getItem('refresh-token');

  if (!accessToken) {
    throw new Error('Missing access token.');
  }

  if (!refreshToken) {
    throw new Error('Missing refresh token.');
  }

  return axios.post<RefreshAccessTokenResponse>('/refresh-token', {
    access_token: accessToken,
    refresh_token: refreshToken,
  });
}
