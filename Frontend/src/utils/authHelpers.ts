export const getTokenData = () => {
  const token = localStorage.getItem('token');
  if (!token) return null;

  const payload = token.split('.')[1];
  if (!payload) return null;

  try {
    const decoded = JSON.parse(atob(payload));
    return decoded;
  } catch (e) {
    console.error("Failed to decode token:", e);
    return null;
  }
};

export const getToken = () => {
  const token = localStorage.getItem('token');
  if(token) {
    return token
  }
  return null
};

export const isManager = () => {
  const tokenData = getTokenData();
  return tokenData?.role === 'Manager';
};

export const isMember = () => {
  const tokenData = getTokenData();
  return tokenData?.role === 'Member';
};

export const hasRole = (roles: string[]) => {
  const tokenData = getTokenData();
  return tokenData ? roles.includes(tokenData.role) : false;
};