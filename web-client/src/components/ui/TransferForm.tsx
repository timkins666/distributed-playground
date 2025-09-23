import { useState } from "react";
import { CurrencyPound } from "@mui/icons-material";
import {
  Button,
  FormControl,
  Grid,
  InputAdornment,
  InputLabel,
  MenuItem,
  OutlinedInput,
  Paper,
  Select,
  Typography,
} from "@mui/material";
import { Account } from "../../types";
import { verifyAndConvertAmount } from "../../utils/currency";

interface TransferFormProps {
  accounts: Account[];
  onTransfer: (from: number, to: number, amount: number) => Promise<void>;
}

export const TransferForm = ({ accounts, onTransfer }: TransferFormProps) => {
  const [transferFrom, setTransferFrom] = useState<string | number>("");
  const [transferTo, setTransferTo] = useState<string | number>("");
  const [transferAmount, setTransferAmount] = useState("");

  const handleTransfer = async () => {
    const amount = verifyAndConvertAmount(transferAmount);
    if (!amount) return;

    await onTransfer(Number(transferFrom), Number(transferTo), amount);
    setTransferFrom("");
    setTransferTo("");
    setTransferAmount("");
  };

  if (accounts.length < 2) return null;

  return (
    <Paper elevation={3} sx={{ p: 3, maxWidth: 600 }}>
      <Typography variant="h6" gutterBottom>
        Transfer Between Accounts
      </Typography>
      <Grid container spacing={2} alignItems="center">
        <Grid size={{ xs: 12, sm: 6 }}>
          <FormControl fullWidth>
            <InputLabel>From</InputLabel>
            <Select
              value={transferFrom}
              label="From"
              onChange={(e) => setTransferFrom(e.target.value)}
            >
              {accounts.map((acc) => (
                <MenuItem key={acc.accountId} value={acc.accountId}>
                  {acc.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <FormControl fullWidth>
            <InputLabel>To</InputLabel>
            <Select
              value={transferTo}
              label="To"
              onChange={(e) => setTransferTo(e.target.value)}
            >
              {accounts.map((acc) => (
                <MenuItem key={acc.accountId} value={acc.accountId}>
                  {acc.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Grid>
        <Grid size={{ xs: 12 }}>
          <OutlinedInput
            fullWidth
            label="Amount"
            type="number"
            value={transferAmount}
            onChange={(e) => setTransferAmount(e.target.value)}
            notched={false}
            startAdornment={
              <InputAdornment position="start">
                <CurrencyPound />
              </InputAdornment>
            }
          />
        </Grid>
        <Grid size={{ xs: 12 }}>
          <Button
            variant="contained"
            fullWidth
            onClick={handleTransfer}
            disabled={
              !transferAmount ||
              !transferTo ||
              !transferFrom ||
              transferFrom === transferTo
            }
          >
            Transfer
          </Button>
        </Grid>
      </Grid>
    </Paper>
  );
};