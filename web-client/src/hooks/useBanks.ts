import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { banksSelector, setBanks, setTriedLoadBanks } from '../store/banksSlice';
import { authSelector } from '../store/authSlice';
import { gatewayUrl } from '../conf';

export const useBanks = () => {
  const dispatch = useDispatch();
  const authStatus = useSelector(authSelector);
  const bankState = useSelector(banksSelector);

  useEffect(() => {
    if (bankState.triedLoad) {
      return;
    }

    dispatch(setTriedLoadBanks());
    const getBanks = async () => {
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
  }, [dispatch, authStatus.token, bankState.triedLoad]);

  return {
    banks: bankState.banks,
  };
};