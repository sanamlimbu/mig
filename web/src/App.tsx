import { Theme } from '@radix-ui/themes';
import Chat from './components/chat';

export default function App() {
  return (
    <Theme className="overflow-y-hidden">
      <Chat />
    </Theme>
  );
}
