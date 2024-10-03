import { DotFilledIcon, DotsVerticalIcon } from '@radix-ui/react-icons';
import { Avatar, Text } from '@radix-ui/themes';
import { getFormattedHourMinute } from '../helpers';
import { Message } from '../types';

interface IMessageProps {
  message: Message;
}

export default function MessageBox({ message }: IMessageProps) {
  const handleAction = () => {};

  return (
    <div className="flex gap-4">
      <Avatar
        src="https://images.unsplash.com/photo-1502823403499-6ccfcf4fb453?&w=256&h=256&q=70&crop=focalpoint&fp-x=0.5&fp-y=0.3&fp-z=1&fit=crop"
        fallback="A"
        radius="full"
      />
      <div>
        <div className="mb-2">
          <Text className="font-semibold">{message.sender_name}</Text>
          <DotFilledIcon className="inline mr-1 ml-1" />
          <Text className="text-gray-500" size="2">
            {getFormattedHourMinute(message.created_at)}
          </Text>
        </div>
        <div className="flex gap-4 items-center">
          <p className="border border-gray-300 rounded-lg p-4 max-w-md">
            {message.content}
          </p>
          <DotsVerticalIcon onClick={handleAction} className="cursor-pointer" />
        </div>
      </div>
    </div>
  );
}
