import { login } from '@/api/auth';
import { Input } from '@/components/ui/input';
import { AuthContext } from '@/contexts/auth';
import { extractErrorMessage } from '@/utils/errors';
import { useMutation } from '@tanstack/react-query';
import { AlertCircle } from 'lucide-react';
import { FormEvent, useContext, useState } from 'react';
import MigIcon96 from '../assets/mig-96.svg';
import { Alert, AlertDescription } from './ui/alert';
import { Button } from './ui/button';
import { CenterDiv } from './ui/center-div';
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
    <CenterDiv className="min-h-screen">
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
        <Alert className="py-2 px-2 text-sm max-w-sm bg-gray-100 mb-4">
          <AlertDescription className="flex">
            <AlertCircle className="h-4 w-4 inline mr-2 mt-1" />
            <p>
              Username: <span className="font-medium">jack</span> and Password:{' '}
              <span className="font-medium">jack123</span> <br />
              Username: <span className="font-medium">jill</span> and Password:{' '}
              <span className="font-medium">jill123</span>
            </p>
          </AlertDescription>
        </Alert>
        {errorMsg && (
          <Alert variant="destructive" className="py-2 px-2 text-sm max-w-sm">
            <AlertDescription>
              <AlertCircle className="h-4 w-4 inline mr-2" />
              {errorMsg}
            </AlertDescription>
          </Alert>
        )}
      </div>
    </CenterDiv>
  );
}
