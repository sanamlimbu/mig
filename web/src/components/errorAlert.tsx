import { AlertCircle } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from './ui/alert';

export default function ErrorAlert({ message }: { message: string }) {
  return (
    <div className="flex items-center justify-center w-full h-full">
      <Alert variant="destructive" className="max-w-sm">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>{message}</AlertDescription>
      </Alert>
    </div>
  );
}
