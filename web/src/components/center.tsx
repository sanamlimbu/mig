import { PropsWithChildren } from 'react';

export default function Center(props: PropsWithChildren) {
  return (
    <div className="flex flex-row min-h-screen justify-center items-center">
      {props.children}
    </div>
  );
}
