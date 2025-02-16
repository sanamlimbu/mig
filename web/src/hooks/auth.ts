import { AuthContext } from '@/contexts/auth';
import { useContext } from 'react';

export function useAuth() {
  const auth = useContext(AuthContext);

  if (!auth) {
    throw new Error('AuthContext is undefined.');
  }

  if (!auth.user) {
    throw new Error('User is null.');
  }

  return {
    user: auth.user,
  };
}
