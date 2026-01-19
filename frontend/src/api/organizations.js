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

// Get team members
export const getTeamMembers = async (orgId, params = {}) => {
  const response = await apiClient.get(`/organizations/${orgId}/members`, { params });
  return response.data;
};

// Invite team member
export const inviteTeamMember = async (orgId, data) => {
  const response = await apiClient.post(`/organizations/${orgId}/members`, data);
  return response.data;
};

// Remove team member
export const removeTeamMember = async (orgId, userId) => {
  const response = await apiClient.delete(`/organizations/${orgId}/members/${userId}`);
  return response.data;
};

// Update team member role
export const updateTeamMemberRole = async (orgId, userId, role) => {
  const response = await apiClient.put(`/organizations/${orgId}/members/${userId}/role`, { role });
  return response.data;
};

// List pending invitations
export const listInvitations = async (orgId) => {
  const response = await apiClient.get(`/organizations/${orgId}/invitations`);
  return response.data;
};

// Revoke invitation
export const revokeInvitation = async (orgId, invitationId) => {
  const response = await apiClient.delete(`/organizations/${orgId}/invitations/${invitationId}`);
  return response.data;
};
