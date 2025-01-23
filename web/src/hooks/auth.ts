import { AuthContext } from '@/contexts/auth';
import { useContext } from 'react';

export function useAuth() {
  const auth = useContext(AuthContext);

  if (!auth) {
    throw new Error('AuthContext is undefined.');
  }

  return {
    isLoggedIn: auth.user ? true : false,
    user: auth.user,
    login: auth.login,
    logout: auth.logout,
  };
}
