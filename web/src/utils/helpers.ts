export const convertDateToFormattedString = (
  str: string,
  includeHourMinute?: boolean
) => {
  const date = new Date(str);

  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');

  return includeHourMinute
    ? `${day}/${month}/${year} ${hours}:${minutes}`
    : `${day}/${month}/${year}`;
};
