import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { setAccounts, setTriedLoadAccounts, updateAccounts, userAccountsSelector } from '../store/accountsSlice';
import { authSelector } from '../store/authSlice';
import { gatewayUrl } from '../conf';

export const useAccounts = () => {
  const dispatch = useDispatch();
  const authStatus = useSelector(authSelector);
  const userAccountState = useSelector(userAccountsSelector);

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
  }, [dispatch, authStatus.token, userAccountState.triedLoad]);

  const createAccount = async (name: string, sourceFundsAccountId?: number, initialBalance?: number) => {
    const res = await fetch(gatewayUrl("account", "new"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${authStatus.token}`,
      },
      body: JSON.stringify({
        name,
        sourceFundsAccountId,
        initialBalance,
      }),
    });
    
    if (res.ok) {
      const data = await res.json();
      dispatch(updateAccounts(data));
      return data;
    }
    throw new Error("Failed to create account");
  };

  const transferFunds = async (sourceAccountId: number, targetAccountId: number, amount: number) => {
    const res = await fetch(gatewayUrl("payment", "transfer"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${authStatus.token}`,
      },
      body: JSON.stringify({
        sourceAccountId,
        targetAccountId,
        appId: crypto.randomUUID(),
        amount,
      }),
    });

    if (res.ok) {
      const updatedAccounts = userAccountState.accounts.map((acc) => {
        if (acc.accountId === sourceAccountId) {
          return { ...acc, balance: acc.balance - amount };
        }
        if (acc.accountId === targetAccountId) {
          return { ...acc, balance: acc.balance + amount };
        }
        return acc;
      });
      dispatch(setAccounts(updatedAccounts));
      return true;
    }
    throw new Error("Failed to transfer funds");
  };

  return {
    accounts: userAccountState.accounts,
    createAccount,
    transferFunds,
  };
};