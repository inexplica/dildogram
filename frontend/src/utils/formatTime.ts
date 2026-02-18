import { format, formatDistanceToNow, isToday, isYesterday, isThisWeek } from 'date-fns';
import { ru } from 'date-fns/locale';

export function formatMessageTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return format(d, 'HH:mm');
}

export function formatMessageDate(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  
  if (isToday(d)) {
    return format(d, 'HH:mm', { locale: ru });
  }
  
  if (isYesterday(d)) {
    return 'вчера в ' + format(d, 'HH:mm', { locale: ru });
  }
  
  if (isThisWeek(d)) {
    return format(d, 'EEEE в HH:mm', { locale: ru });
  }
  
  return format(d, 'dd.MM.yyyy в HH:mm', { locale: ru });
}

export function formatLastSeen(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return formatDistanceToNow(d, { addSuffix: true, locale: ru });
}

export function formatChatTime(date: string | Date | null): string {
  if (!date) return '';
  
  const d = typeof date === 'string' ? new Date(date) : date;
  
  if (isToday(d)) {
    return format(d, 'HH:mm');
  }
  
  if (isYesterday(d)) {
    return 'вчера';
  }
  
  if (isThisWeek(d)) {
    return format(d, 'EEE');
  }
  
  return format(d, 'dd.MM.yy');
}
