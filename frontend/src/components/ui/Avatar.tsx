import React from 'react';
import clsx from 'clsx';

interface AvatarProps {
  src?: string;
  alt: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
  online?: boolean;
}

const sizeClasses = {
  sm: 'w-8 h-8 text-xs',
  md: 'w-10 h-10 text-sm',
  lg: 'w-12 h-12 text-base',
  xl: 'w-16 h-16 text-lg',
};

export function Avatar({ src, alt, size = 'md', className, online }: AvatarProps) {
  const initials = alt
    .split(' ')
    .map((n) => n[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();

  const bgColor = stringToColor(alt);

  return (
    <div className={clsx('relative inline-block', className)}>
      {src ? (
        <img
          src={src}
          alt={alt}
          className={clsx(
            'rounded-full object-cover',
            sizeClasses[size]
          )}
        />
      ) : (
        <div
          className={clsx(
            'rounded-full flex items-center justify-center font-medium text-white',
            sizeClasses[size]
          )}
          style={{ backgroundColor: bgColor }}
        >
          {initials}
        </div>
      )}
      {online && (
        <span className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-chat-bg rounded-full" />
      )}
    </div>
  );
}

function stringToColor(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  
  const hue = hash % 360;
  return `hsl(${hue}, 65%, 45%)`;
}
