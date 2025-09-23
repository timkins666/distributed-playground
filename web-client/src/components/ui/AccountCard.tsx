import { Card, CardContent, Typography } from "@mui/material";
import { Account } from "../../types";
import { formatCurrency } from "../../utils/currency";

interface AccountCardProps {
  account: Account;
}

export const AccountCard = ({ account }: AccountCardProps) => {
  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="h6">{account.name}</Typography>
        <Typography color="text.secondary">
          Balance: {formatCurrency(account.balance)}
        </Typography>
      </CardContent>
    </Card>
  );
};