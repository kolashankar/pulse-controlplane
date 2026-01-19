import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useNavigate } from 'react-router-dom';
import { listProjects } from '@/api/projects';
import { listOrganizations } from '@/api/organizations';
import { getRecentLogs } from '@/api/auditLogs';
import { Boxes, Building2, Users, Activity, Plus, ArrowRight } from 'lucide-react';
import { toast } from 'sonner';

const Dashboard = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    projects: 0,
    organizations: 0,
    teamMembers: 0
  });
  const [recentActivity, setRecentActivity] = useState([]);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      const [projectsRes, orgsRes, activityRes] = await Promise.all([
        listProjects({ page: 1, limit: 1 }),
        listOrganizations({ page: 1, limit: 1 }),
        getRecentLogs({ limit: 5 })
      ]);

      setStats({
        projects: projectsRes.total || 0,
        organizations: orgsRes.total || 0,
        teamMembers: 0 // Will be updated when team API is called
      });

      setRecentActivity(activityRes.logs || activityRes || []);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      toast.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const StatCard = ({ title, value, icon: Icon, onClick }) => (
    <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={onClick}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
      </CardContent>
    </Card>
  );

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-64" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map(i => <Skeleton key={i} className="h-32" />)}
        </div>
        <Skeleton className="h-96" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">
            Welcome to Pulse Control Plane
          </p>
        </div>
        <Button onClick={() => navigate('/projects/new')}>
          <Plus className="mr-2 h-4 w-4" />
          Create Project
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Projects"
          value={stats.projects}
          icon={Boxes}
          onClick={() => navigate('/projects')}
        />
        <StatCard
          title="Organizations"
          value={stats.organizations}
          icon={Building2}
          onClick={() => navigate('/organizations')}
        />
        <StatCard
          title="Team Members"
          value={stats.teamMembers}
          icon={Users}
          onClick={() => navigate('/team')}
        />
        <StatCard
          title="System Status"
          value="Operational"
          icon={Activity}
          onClick={() => navigate('/status')}
        />
      </div>

      {/* Recent Activity */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>Latest actions in your organization</CardDescription>
          </CardHeader>
          <CardContent>
            {recentActivity.length > 0 ? (
              <div className="space-y-4">
                {recentActivity.slice(0, 5).map((log, idx) => (
                  <div key={idx} className="flex items-start gap-3 text-sm">
                    <Activity className="h-4 w-4 mt-0.5 text-muted-foreground" />
                    <div className="flex-1">
                      <p className="font-medium">{log.action}</p>
                      <p className="text-muted-foreground text-xs">
                        {log.user_email} • {new Date(log.timestamp).toLocaleString()}
                      </p>
                    </div>
                  </div>
                ))}
                <Button
                  variant="outline"
                  className="w-full mt-4"
                  onClick={() => navigate('/audit-logs')}
                >
                  View All Logs
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No recent activity</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common tasks and shortcuts</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button variant="outline" className="w-full justify-start" onClick={() => navigate('/projects/new')}>
              <Plus className="mr-2 h-4 w-4" />
              Create New Project
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => navigate('/team')}>
              <Users className="mr-2 h-4 w-4" />
              Invite Team Member
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => navigate('/billing')}>
              <Activity className="mr-2 h-4 w-4" />
              View Usage & Billing
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => navigate('/status')}>
              <Activity className="mr-2 h-4 w-4" />
              Check System Status
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Dashboard;
