import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Login from './components/login';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Login />
    </QueryClientProvider>
  );
}

export default App;
