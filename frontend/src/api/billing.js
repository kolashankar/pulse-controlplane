import apiClient from './client';

// Get billing dashboard
export const getBillingDashboard = async (projectId) => {
  const response = await apiClient.get(`/billing/${projectId}/dashboard`);
  return response.data;
};

// Generate invoice
export const generateInvoice = async (projectId, data) => {
  const response = await apiClient.post(`/billing/${projectId}/invoice`, data);
  return response.data;
};

// Get invoice
export const getInvoice = async (invoiceId) => {
  const response = await apiClient.get(`/billing/invoice/${invoiceId}`);
  return response.data;
};

// List invoices
export const listInvoices = async (projectId, params = {}) => {
  const response = await apiClient.get(`/billing/${projectId}/invoices`, { params });
  return response.data;
};

// Update invoice status
export const updateInvoiceStatus = async (invoiceId, status) => {
  const response = await apiClient.put(`/billing/invoice/${invoiceId}/status`, { status });
  return response.data;
};
