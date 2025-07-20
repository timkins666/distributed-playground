import {
    Box,
    Button,
    Card,
    CardContent,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    Grid,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    TextField,
    Typography,
} from "@mui/material";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
    addAccount,
    setAccounts,
    setTriedLoadAccounts,
    userAccountsSelector,
} from "../components/accounts/accountsSlice";
import { authSelector } from "../components/auth/authSlice";
import {
    banksSelector,
    setBanks,
    setTriedLoadBanks,
} from "../components/banks/banksSlice";
import { gatewayUrl } from "../conf";

const UserDashboard = () => {
  const dispatch = useDispatch();
  const authStatus = useSelector(authSelector);
  const bankState = useSelector(banksSelector);
  const userAccountState = useSelector(userAccountsSelector);
  const accounts = userAccountState.accounts;

  useEffect(() => {
    if (bankState.triedLoad) {
      console.log("Already tried loading banks, not trying again");
    }

    dispatch(setTriedLoadBanks());
    const getBanks = async () => {
      if (bankState.triedLoad) {
        return;
      }
      const res = await fetch(gatewayUrl("account", "banks"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authStatus.token}`,
        },
      });
      const data = await res.json();
      dispatch(setBanks(data));
    };
    getBanks();
  }, []);

  useEffect(() => {
    if (userAccountState.triedLoad) {
      return;
    }

    dispatch(setTriedLoadAccounts());
    const getAccounts = async () => {
      const res = await fetch(gatewayUrl("account", "myaccounts"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authStatus.token}`,
        },
      });
      const data = await res.json();
      dispatch(setAccounts(data));
    };
    getAccounts();
  }, []);

  const [createOpen, setCreateOpen] = useState(false);
  const [newAccountSourceId, setNewAccountSourceId] = useState("");
  const [newAccountInitialBalance, setNewAccountInitialBalance] = useState("");
  const [newAccountName, setNewAccountName] = useState("");

  const [transferFrom, setTransferFrom] = useState("");
  const [transferTo, setTransferTo] = useState("");
  const [transferAmount, setTransferAmount] = useState("");

  const handleCreateAccount = async () => {
    const res = await fetch(gatewayUrl("account", "new"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${authStatus.token}`,
    },
    body: JSON.stringify({name: newAccountName}),
    });
    const data = await res.json();
    dispatch(addAccount(data));
    setCreateOpen(false);
  };

  const handleTransfer = async () => {
    let amount = parseFloat(transferAmount);

    if (isNaN(amount)) {
      console.error("amount not a number");
      return; // do error
    }

    amount *= 100;
    if (amount <= 0 || amount.toString().includes(".")) {
      console.error("invalid amount");
      return; // do error
    }

    const res = await fetch(gatewayUrl("payment", "pay"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${authStatus.token}`,
      },
      body: JSON.stringify({
        sourceAccountId: transferFrom,
        targetAccountId: transferTo,
        appId: Date.now().toString(),
        amount: amount,
      }),
    });
    const success = res.status === 202;
    console.log("success: ", success);
  };

  return (
    <Box sx={{ padding: 4 }}>
      <Typography variant="h4" gutterBottom>
        My Accounts
      </Typography>

      <Grid container spacing={2} mb={4}>
        {accounts.map((account) => (
          <Grid size={{ xs: 12, md: 6, lg: 4 }} key={account.accountId}>
            <Card variant="outlined">
              <CardContent>
                <Typography variant="h6">{account.name}</Typography>
                <Typography color="text.secondary">
                  Balance: £{account.balance / 100}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Button
        variant="contained"
        onClick={() => setCreateOpen(true)}
        sx={{ mb: 4 }}
      >
        Create New Account
      </Button>

      {/* Transfer Section */}
      {accounts.length >= 2 && (
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
              <TextField
                fullWidth
                label="Amount"
                type="number"
                value={transferAmount}
                onChange={(e) => setTransferAmount(e.target.value)}
              />
            </Grid>
            <Grid size={{ xs: 12 }}>
              <Button variant="contained" fullWidth onClick={handleTransfer}>
                Transfer
              </Button>
            </Grid>
          </Grid>
        </Paper>
      )}

      {/* Create New Account Dialog */}
      <Dialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        maxWidth="sm"
        fullWidth
      >
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
              No existing accounts — a new account will credit you with a random
              amount
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button onClick={handleCreateAccount} variant="contained">
            Create
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default UserDashboard;
