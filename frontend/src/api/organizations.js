import apiClient from './client';

// Create organization
export const createOrganization = async (data) => {
  const response = await apiClient.post('/organizations', data);
  return response.data;
};

// List organizations
export const listOrganizations = async (params = {}) => {
  const response = await apiClient.get('/organizations', { params });
  return response.data;
};

// Get organization by ID
export const getOrganization = async (id) => {
  const response = await apiClient.get(`/organizations/${id}`);
  return response.data;
};

// Update organization
export const updateOrganization = async (id, data) => {
  const response = await apiClient.put(`/organizations/${id}`, data);
  return response.data;
};

// Delete organization
export const deleteOrganization = async (id) => {
  const response = await apiClient.delete(`/organizations/${id}`);
  return response.data;
};
