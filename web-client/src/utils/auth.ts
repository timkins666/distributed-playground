import dayjs from 'dayjs';
import { UserCreds } from '../types';

const LOCAL_STORAGE_KEY = 'login';

export const parseToken = (token: string | null) => {
  try {
    const payload = JSON.parse(atob(token?.split('.')[1] ?? '')) as {
      exp: number;
      iat: number;
    };
    if (!payload.exp) {
      return null;
    }
    if (Date.now() >= payload.exp * 1000) {
      console.log('Token expired');
      clearLocalStorageLogin();
      return null;
    }

    console.log(
      `Reusing token (expires ${dayjs(payload.exp * 1000).format(
        'DD/MM/YYYY HH:mm:ss'
      )})`
    );
    return payload;
  } catch {
    return null;
  }
};

export const getLocalStorageLogin = () => {
  const saved = localStorage.getItem(LOCAL_STORAGE_KEY);

  if (!saved) {
    return null;
  }

  const creds: UserCreds = JSON.parse(saved);
  const token = parseToken(creds.token);

  if (!token) {
    return null;
  }

  return creds;
};

export const setLocalStorageLogin = (login: UserCreds) => {
  localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(login));
};

export const clearLocalStorageLogin = () => {
  localStorage.removeItem(LOCAL_STORAGE_KEY);
};
