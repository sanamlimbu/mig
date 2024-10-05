export function getFormattedHourMinute(time: Date) {
  return time.toLocaleString('en-US', {
    hour: 'numeric',
    minute: 'numeric',
    hour12: true,
  });
}
