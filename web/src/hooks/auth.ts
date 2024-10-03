import { Session } from '@supabase/supabase-js';
import { useContext } from 'react';
import { SessionContext } from '../providers/session';

export function useAuth() {
  const session = useContext(SessionContext);

  if (!session) {
    return {
      user: null,
      isLoggedIn: false,
    };
  }

  return {
    user: getUserFromSession(session),
    isLoggedIn: true,
  };
}

const getUserFromSession = (session: Session) => {
  return {
    id: session.user.id,
    email: session.user.email || null,
    first_name: session.user.user_metadata['first_name'] || null,
    last_name: session.user.user_metadata['last_name'] || null,
    app_role: session.user.app_metadata['app_role'] || null,
    created_at: session.user.created_at,
    last_sign_in_at: session.user.last_sign_in_at || '',
  };
};
