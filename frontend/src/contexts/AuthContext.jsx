import React, { createContext, useContext, useState, useEffect } from 'react';

const AuthContext = createContext(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [currentOrg, setCurrentOrg] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Load organization from localStorage on mount
    const savedOrg = localStorage.getItem('pulse_current_org');
    if (savedOrg) {
      try {
        setCurrentOrg(JSON.parse(savedOrg));
      } catch (e) {
        console.error('Failed to parse saved org:', e);
      }
    }
    setLoading(false);
  }, []);

  const selectOrganization = (org) => {
    setCurrentOrg(org);
    localStorage.setItem('pulse_current_org', JSON.stringify(org));
  };

  const clearOrganization = () => {
    setCurrentOrg(null);
    localStorage.removeItem('pulse_current_org');
  };

  const value = {
    currentOrg,
    setCurrentOrg: selectOrganization,
    clearOrganization,
    isAuthenticated: !!currentOrg,
    loading
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
