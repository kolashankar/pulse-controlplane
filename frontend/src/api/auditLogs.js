import apiClient from './client';

// Get audit logs
export const getAuditLogs = async (params = {}) => {
  const response = await apiClient.get('/audit-logs', { params });
  return response.data;
};

// Export audit logs
export const exportAuditLogs = async (params = {}) => {
  const response = await apiClient.get('/audit-logs/export', { 
    params,
    responseType: 'blob'
  });
  return response.data;
};

// Get audit stats
export const getAuditStats = async () => {
  const response = await apiClient.get('/audit-logs/stats');
  return response.data;
};

// Get recent logs
export const getRecentLogs = async (params = {}) => {
  const response = await apiClient.get('/audit-logs/recent', { params });
  return response.data;
};
