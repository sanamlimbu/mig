import { AuthToken } from '@/api/auth';
import { User } from '@/types';
import { createContext } from 'react';

interface AuthContextValue {
  user: User | null;
  login: (data: AuthToken) => void;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextValue | undefined>(
  undefined
);
