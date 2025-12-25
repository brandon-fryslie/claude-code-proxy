import { type FC } from 'react';
import { cn } from '@/lib/utils';

interface MessageContentProps {
  content: string;
  className?: string;
}

export const MessageContent: FC<MessageContentProps> = ({ content, className }) => {
  return (
    <div className={cn('prose prose-sm max-w-none', className)}>
      <p className="whitespace-pre-wrap text-gray-700">{content}</p>
    </div>
  );
};
