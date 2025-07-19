import { Button } from "@mui/material";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";

import {
  addAccount,
  setAccounts,
  setTriedLoad as triedLoadAccounts,
  userAccountsSelector,
} from "../components/accounts/accountsSlice";
import { authSelector } from "../components/auth/authSlice";
import {
  banksSelector,
  setBanks,
  setTriedLoad as triedLoadBanks,
} from "../components/banks/banksSlice";
import { gatewayUrl } from "../conf";

export default function Accounts() {
  const dispatch = useDispatch();
  const authStatus = useSelector(authSelector);
  const bankState = useSelector(banksSelector);
  const userAccountState = useSelector(userAccountsSelector);

  useEffect(() => {
    if (bankState.triedLoad) {
      console.log("Already tried loading banks, not trying again");
    }

    dispatch(triedLoadBanks());
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

    dispatch(triedLoadAccounts());
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

  const createAccount = async () => {
    const res = await fetch(gatewayUrl("account", "new"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${authStatus.token}`,
        // body: JSON.stringify({})
      },
    });
    const data = await res.json();
    dispatch(addAccount(data));
  };

  return (
    <div>
      <h2>Accounts</h2>
      <h5>Banks</h5>
      {bankState.banks.map((bank) => (
        <p key={bank.id}>{bank.name}</p>
      ))}
      <h5>User accounts</h5>
      {!userAccountState.accounts.length ? (
        <p>You have no accounts :/</p>
      ) : (
        userAccountState.accounts.map((acc) => (
          <p key={acc.accountId}>
            {acc.bankName}: {acc.balance}
          </p>
        ))
      )}

      <Button onClick={createAccount}>Create new account</Button>
    </div>
  );
}
