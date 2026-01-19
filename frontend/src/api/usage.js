import apiClient from './client';

// Get usage metrics
export const getUsageMetrics = async (projectId, params = {}) => {
  const response = await apiClient.get(`/usage/${projectId}`, { params });
  return response.data;
};

// Get usage summary
export const getUsageSummary = async (projectId, params = {}) => {
  const response = await apiClient.get(`/usage/${projectId}/summary`, { params });
  return response.data;
};

// Get aggregated usage
export const getAggregatedUsage = async (projectId, params = {}) => {
  const response = await apiClient.get(`/usage/${projectId}/aggregated`, { params });
  return response.data;
};

// Get usage alerts
export const getUsageAlerts = async (projectId) => {
  const response = await apiClient.get(`/usage/${projectId}/alerts`);
  return response.data;
};

// Check usage limits
export const checkUsageLimits = async (projectId) => {
  const response = await apiClient.post(`/usage/${projectId}/check-limits`);
  return response.data;
};
