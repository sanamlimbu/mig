import { PropsWithChildren } from 'react';

export const CenterDiv = (props: PropsWithChildren) => {
  return (
    <div className="w-full h-full flex flex-row justify-center items-center">
      {props.children}
    </div>
  );
};
