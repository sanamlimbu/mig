import { AuthContext } from '@/contexts/auth';
import { getAuthToken } from '@/utils/auth';
import { useContext } from 'react';

export function useAuth() {
  const auth = useContext(AuthContext);
  const authToken = getAuthToken();

  if (!authToken) {
    throw new Error('Auth token is missing.');
  }

  if (!auth) {
    throw new Error('AuthContext is undefined.');
  }

  if (!auth.user) {
    throw new Error('User is null.');
  }

  return {
    user: auth.user,
    accessToken: authToken.access_token,
  };
}
