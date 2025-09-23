import { Button } from '@mui/material';
import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router';
import { authSelector, logout } from '../store/authSlice';
import { clearLocalStorageLogin } from '../utils/auth';

const DELAY_FOR_NO_REASON=2

export default function Logout() {
  const authStatus = useSelector(authSelector);
  const dispatch = useDispatch();

  useEffect(() => {
    const doLogout = () => {
      const timer = setTimeout(() => {
        dispatch(logout());
        clearLocalStorageLogin();
      }, DELAY_FOR_NO_REASON * 1000);

      return () => {
        clearTimeout(timer);
      };
    };
    doLogout();
  }, []);

  if (authStatus.loggedIn) {
    return <>logging out in {DELAY_FOR_NO_REASON} seconds...</>;
  } else {
    return (
      <>
        <p>logged out</p>
        <Button>
          <Link to='/login'>Return to login</Link>
        </Button>
      </>
    );
  }
}
