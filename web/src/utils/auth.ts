import { AuthToken } from '@/api/auth';

export const setAuthToken = (data: AuthToken) => {
  sessionStorage.setItem('auth-token', JSON.stringify(data));
};

export const getAuthToken = () => {
  const value = sessionStorage.getItem('auth-token');
  return value ? (JSON.parse(value) as AuthToken) : null;
};

export const removeAuthToken = () => {
  sessionStorage.removeItem('auth-token');
};
