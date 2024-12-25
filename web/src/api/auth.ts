import { User } from '@/types';
import { HttpStatusCode } from 'axios';
import { axios } from '../axios';

interface LoginResponse {
  access_token: string;
  expires_at: number;
  expires_in: string;
  refresh_token: string;
  user: User;
}

export async function login(username: string, password: string) {
  try {
    const { data, status } = await axios.post(
      '/login',
      {
        username: username,
        password: password,
      },
      {
        headers: {
          'Content-Type': 'application/json',
        },
      }
    );

    if (status !== HttpStatusCode.Ok) {
      throw new Error('Invalid username or password.');
    }

    return data as LoginResponse;
  } catch (err) {
    throw err as Error;
  }
}
