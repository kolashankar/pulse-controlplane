import apiClient from './client';

// Create token
export const createToken = async (data) => {
  const response = await apiClient.post('/tokens/create', data);
  return response.data;
};

// Validate token
export const validateToken = async (token) => {
  const response = await apiClient.post('/tokens/validate', { token });
  return response.data;
};
