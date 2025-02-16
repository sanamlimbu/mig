import { useContext } from 'react';
import Home from './components/home';
import Login from './components/login';
import { AuthContext } from './contexts/auth';

export default function App() {
  const auth = useContext(AuthContext);
  if (auth?.user === null) {
    return <Login />;
  }

  return <Home />;
}
