import apiClient from './client';

// Get system status
export const getSystemStatus = async () => {
  const response = await apiClient.get('/status');
  return response.data;
};

// Get project health
export const getProjectHealth = async (projectId) => {
  const response = await apiClient.get(`/status/projects/${projectId}`);
  return response.data;
};

// Get region availability
export const getRegionAvailability = async () => {
  const response = await apiClient.get('/status/regions');
  return response.data;
};
