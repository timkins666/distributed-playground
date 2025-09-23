import { useState } from "react";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  Typography,
} from "@mui/material";
import { Account } from "../../types";
import { verifyAndConvertAmount } from "../../utils/currency";

interface CreateAccountDialogProps {
  open: boolean;
  onClose: () => void;
  accounts: Account[];
  onCreateAccount: (name: string, sourceFundsAccountId?: number, initialBalance?: number) => Promise<void>;
}

export const CreateAccountDialog = ({
  open,
  onClose,
  accounts,
  onCreateAccount,
}: CreateAccountDialogProps) => {
  const [newAccountSourceId, setNewAccountSourceId] = useState("");
  const [newAccountInitialBalance, setNewAccountInitialBalance] = useState("");
  const [newAccountName, setNewAccountName] = useState("");

  const handleCreateAccount = async () => {
    let initialBalance: number | null = 0;
    if (accounts.length > 0) {
      initialBalance = verifyAndConvertAmount(newAccountInitialBalance);
      if (initialBalance === null) return;
    }

    await onCreateAccount(
      newAccountName,
      newAccountSourceId ? parseInt(newAccountSourceId) : undefined,
      initialBalance || undefined
    );
    
    setNewAccountName("");
    setNewAccountSourceId("");
    setNewAccountInitialBalance("");
    onClose();
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Create New Account</DialogTitle>
      <DialogContent>
        <FormControl fullWidth sx={{ mt: 2 }} required>
          <InputLabel>Name your account</InputLabel>
          <TextField
            fullWidth
            sx={{ mt: 2 }}
            value={newAccountName}
            onChange={(e) => setNewAccountName(e.target.value)}
          />
        </FormControl>
        {accounts.length > 0 ? (
          <>
            <FormControl fullWidth sx={{ mt: 2 }}>
              <InputLabel>Transfer initial balance from</InputLabel>
              <Select
                value={newAccountSourceId}
                label="Transfer From"
                onChange={(e) => setNewAccountSourceId(e.target.value)}
              >
                {accounts.map((acc) => (
                  <MenuItem key={acc.accountId} value={acc.accountId}>
                    {acc.name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              fullWidth
              label="Initial Balance"
              type="number"
              sx={{ mt: 2 }}
              value={newAccountInitialBalance}
              onChange={(e) => setNewAccountInitialBalance(e.target.value)}
            />
          </>
        ) : (
          <Typography>
            No existing accounts — a new account will credit you with a random amount
          </Typography>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleCreateAccount} variant="contained">
          Create
        </Button>
      </DialogActions>
    </Dialog>
  );
};