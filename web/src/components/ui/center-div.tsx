import { ReactNode } from 'react';

export const CenterDiv = (props: {
  className?: string;
  children: ReactNode;
}) => {
  return (
    <div
      className={`${props.className} w-full h-full flex flex-row justify-center items-center`}
    >
      {props.children}
    </div>
  );
};
