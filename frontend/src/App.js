import React from "react";
import "@/App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import Layout from "@/components/Layout";

// Pages
import Dashboard from "@/pages/Dashboard";
import Organizations from "@/pages/Organizations";
import Projects from "@/pages/Projects";
import ProjectDetails from "@/pages/ProjectDetails";
import Billing from "@/pages/Billing";
import Team from "@/pages/Team";
import AuditLogs from "@/pages/AuditLogs";
import Status from "@/pages/Status";
import ChatMessaging from "@/pages/ChatMessaging";
import VideoAudio from "@/pages/VideoAudio";
import Moderation from "@/pages/Moderation";

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Redirect root to dashboard */}
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          
          {/* Main routes with layout */}
          <Route path="/dashboard" element={<Layout><Dashboard /></Layout>} />
          <Route path="/organizations" element={<Layout><Organizations /></Layout>} />
          <Route path="/projects" element={<Layout><Projects /></Layout>} />
          <Route path="/projects/:id" element={<Layout><ProjectDetails /></Layout>} />
          <Route path="/billing" element={<Layout><Billing /></Layout>} />
          <Route path="/team" element={<Layout><Team /></Layout>} />
          <Route path="/audit-logs" element={<Layout><AuditLogs /></Layout>} />
          <Route path="/status" element={<Layout><Status /></Layout>} />
          
          {/* Feature routes */}
          <Route path="/features/chat" element={<Layout><ChatMessaging /></Layout>} />
          <Route path="/features/video" element={<Layout><VideoAudio /></Layout>} />
          <Route path="/features/moderation" element={<Layout><Moderation /></Layout>} />
          
          {/* Catch all - redirect to dashboard */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
