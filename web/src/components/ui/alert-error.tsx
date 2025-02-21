import { AlertCircle } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from './alert';
import { CenterDiv } from './center-div';

export const AlertError = (props: { title: string; message: string }) => {
  return (
    <CenterDiv>
      <Alert variant="destructive" className="max-w-md">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>{props.title}</AlertTitle>
        <AlertDescription>{props.message}</AlertDescription>
      </Alert>
    </CenterDiv>
  );
};
