import { Input } from './ui/input';
import { ScrollArea } from './ui/scroll-area';

export default function Chatrooms() {
  const handleSearchTermChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    console.log(e);
  };

  return (
    <div>
      <h1 className="font-bold">Chats</h1>
      <Input
        type="text"
        className="my-3"
        onChange={handleSearchTermChange}
        placeholder="Search"
      />
      <ScrollArea className="max-w-sm h-75">
        Jokester began sneaking into the castle in the middle of the night and
        leaving jokes all over the place: under the king's pillow, in his soup,
        even in the royal toilet. The king was furious, but he couldn't seem to
        stop Jokester. And then, one day, the people of the kingdom discovered
        that the jokes left by Jokester were so funny that they couldn't help
        but laugh. And once they started laughing, they couldn't stop.
      </ScrollArea>
    </div>
  );
}
