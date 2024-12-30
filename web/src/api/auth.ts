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
