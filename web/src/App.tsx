import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Home from './components/home';
import AuthProvider from './providers/auth';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Home />
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
