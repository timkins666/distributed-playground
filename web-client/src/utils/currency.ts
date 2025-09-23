export const verifyAndConvertAmount = (amount: number | string): number | null => {
  amount = parseFloat(amount as string);

  if (isNaN(amount)) {
    console.error("amount not a number");
    return null;
  }

  amount = amount * 1000 / 10;
  if (amount <= 0 || amount.toString().includes(".")) {
    console.error("invalid amount");
    return null;
  }

  return amount;
};

export const formatCurrency = (amount: number): string => {
  return `£${amount / 100}`;
};