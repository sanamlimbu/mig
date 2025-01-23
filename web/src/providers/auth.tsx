import { AuthToken } from '@/api/auth';
import { AuthContext } from '@/contexts/auth';
import { User } from '@/types';
import { getAuthToken, removeAuthToken, setAuthToken } from '@/utils/auth';
import { PropsWithChildren, useEffect, useState } from 'react';

export default function AuthProvider(props: PropsWithChildren) {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const syncLogout = (event: StorageEvent) => {
      if (event.key === 'logout') {
        removeAuthToken();
        setUser(null);
      }
    };

    window.addEventListener('storage', syncLogout);

    return window.removeEventListener('storage', syncLogout);
  }, []);

  useEffect(() => {
    const syncSession = (event: StorageEvent) => {
      if (event.key === 'get-session-storage') {
        localStorage.setItem('session-storage', JSON.stringify(sessionStorage));
        localStorage.removeItem('session-storage');
      } else if (event.key === 'session-storage' && !sessionStorage.length) {
        const newValue = event.newValue;
        if (!newValue) {
          throw new Error(
            'Missing event value for "session-storage" event key.'
          );
        }

        const data = JSON.parse(newValue);
        for (const key in data) {
          sessionStorage.setItem(key, data[key]);
        }

        const authToken = getAuthToken();
        setUser(authToken ? authToken.user : null);
      }
    };

    window.addEventListener('storage', syncSession);

    if (!sessionStorage.length) {
      localStorage.setItem('get-session-storage', String(Date.now()));
    }

    return window.removeEventListener('storage', syncSession);
  }, []);

  const login = (data: AuthToken) => {
    setAuthToken(data);
    setUser(data.user);
  };

  const logout = () => {
    // This will trigger an event and logout operations are handled by event listener.
    localStorage.setItem('logout', String(Date.now()));
  };

  return (
    <AuthContext.Provider value={{ user, login, logout }}>
      {props.children}
    </AuthContext.Provider>
  );
}
