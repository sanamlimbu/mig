export const convetDateToFormattedString = (str: string) => {
  const date = new Date(str);

  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();

  return `${day}/${month}/${year}`;
};
