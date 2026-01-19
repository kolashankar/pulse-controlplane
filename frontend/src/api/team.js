import apiClient from './client';

// List team members
export const listTeamMembers = async (orgId, params = {}) => {
  const response = await apiClient.get(`/organizations/${orgId}/members`, { params });
  return response.data;
};

// Invite team member
export const inviteTeamMember = async (orgId, data) => {
  const response = await apiClient.post(`/organizations/${orgId}/members`, data);
  return response.data;
};

// Get team member
export const getTeamMember = async (orgId, userId) => {
  const response = await apiClient.get(`/organizations/${orgId}/members/${userId}`);
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

// Accept invitation
export const acceptInvitation = async (token) => {
  const response = await apiClient.post('/invitations/accept', { token });
  return response.data;
};
