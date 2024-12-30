import { AuthContext } from '@/contexts/auth';
import { User } from '@/types';
import { PropsWithChildren, useState } from 'react';

export default function AuthProvider(props: PropsWithChildren) {
  const token = localStorage.getItem('mig-auth-token');
  const [user, setUser] = useState<User | null>(
    token ? JSON.parse(token).user : null
  );

  const login = (user: User) => {
    setUser(user);
  };

  const logout = () => {
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout }}>
      {props.children}
    </AuthContext.Provider>
  );
}
