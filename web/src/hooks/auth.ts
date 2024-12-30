import { AuthContext } from '@/contexts/auth';
import { useContext } from 'react';

export function useAuth() {
  const auth = useContext(AuthContext);

  if (auth === undefined) {
    throw new Error('AuthContext is undefined.');
  }

  return {
    user: auth.user,
    login: auth.login,
    logout: auth.logout,
  };
}
