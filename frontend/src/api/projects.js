import apiClient from './client';

// Create project
export const createProject = async (data) => {
  const response = await apiClient.post('/projects', data);
  return response.data;
};

// List projects
export const listProjects = async (params = {}) => {
  const response = await apiClient.get('/projects', { params });
  return response.data;
};

// Get project by ID
export const getProject = async (id) => {
  const response = await apiClient.get(`/projects/${id}`);
  return response.data;
};

// Update project
export const updateProject = async (id, data) => {
  const response = await apiClient.put(`/projects/${id}`, data);
  return response.data;
};

// Delete project
export const deleteProject = async (id) => {
  const response = await apiClient.delete(`/projects/${id}`);
  return response.data;
};

// Regenerate API keys
export const regenerateAPIKeys = async (id) => {
  const response = await apiClient.post(`/projects/${id}/regenerate-keys`);
  return response.data;
};
