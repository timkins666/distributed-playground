import { Box, Button, Grid, Typography } from "@mui/material";
import { useState } from "react";
import { AccountCard } from "../components/ui/AccountCard";
import { CreateAccountDialog } from "../components/ui/CreateAccountDialog";
import { TransferForm } from "../components/ui/TransferForm";
import { useAccounts } from "../hooks/useAccounts";

const AccountsPage = () => {
  const { accounts, createAccount, transferFunds } = useAccounts();
  const [createOpen, setCreateOpen] = useState(false);

  const handleCreateAccount = async (name: string, sourceFundsAccountId?: number, initialBalance?: number) => {
    await createAccount(name, sourceFundsAccountId, initialBalance);
  };

  const handleTransfer = async (from: number, to: number, amount: number) => {
    await transferFunds(from, to, amount);
  };

  return (
    <Box sx={{ padding: 4 }}>
      <Typography variant="h4" gutterBottom>
        My Accounts
      </Typography>

      <Grid container spacing={2} mb={4}>
        {[...accounts].sort((a,b) => a.name < b.name ? -1 : 1).map((account) => (
          <Grid size={{ xs: 12, md: 6, lg: 4 }} key={account.accountId}>
            <AccountCard account={account}/>
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

      <TransferForm accounts={accounts} onTransfer={handleTransfer} />

      <CreateAccountDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        accounts={accounts}
        onCreateAccount={handleCreateAccount}
      />
    </Box>
  );
};

export default AccountsPage;