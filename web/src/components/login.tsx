import { login } from '@/api/auth';
import { Input } from '@/components/ui/input';
import { AuthContext } from '@/contexts/auth';
import { extractErrorMessage } from '@/utils/errors';
import { useMutation } from '@tanstack/react-query';
import { AlertCircle } from 'lucide-react';
import { FormEvent, useContext, useState } from 'react';
import MigIcon96 from '../assets/mig-96.svg';
import Center from './center';
import { Alert, AlertDescription } from './ui/alert';
import { Button } from './ui/button';
import { Label } from './ui/label';

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const auth = useContext(AuthContext);

  const mutation = useMutation({
    mutationFn: () => login(username, password),
    onSuccess: (data) => {
      auth?.login(data);
    },
    onError: (error) => {
      setErrorMsg(extractErrorMessage(error));
    },
  });

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setErrorMsg('');

    if (username === '') {
      setErrorMsg('Username is required.');
      return;
    }

    if (password === '') {
      setErrorMsg('Password is required.');
      return;
    }

    mutation.mutate();
  };

  return (
    <Center>
      <div className="w-full max-w-sm">
        <img src={MigIcon96} className="mb-4" />
        <form onSubmit={handleSubmit} className="max-w-md">
          <Label>Username </Label>
          <Input
            type="text"
            placeholder="Username"
            onChange={(e) => {
              setErrorMsg('');
              setUsername(e.currentTarget.value);
            }}
            className="mb-4"
          />
          <Label>Password</Label>
          <Input
            type="password"
            placeholder="Password"
            onChange={(e) => {
              setErrorMsg('');
              setPassword(e.currentTarget.value);
            }}
            className="mb-4"
          />
          <Button variant="outline" type="submit" className="mb-4">
            Login
          </Button>
        </form>
        {errorMsg && (
          <Alert variant="destructive" className="py-2 px-2 text-sm max-w-sm">
            <AlertDescription>
              <AlertCircle className="h-4 w-4 inline mr-2" />
              {errorMsg}
            </AlertDescription>
          </Alert>
        )}
      </div>
    </Center>
  );
}
