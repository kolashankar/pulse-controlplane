import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { CheckCircle2, XCircle, AlertCircle, Activity } from 'lucide-react';
import { getSystemStatus, getRegionAvailability } from '@/api/status';
import { toast } from 'sonner';

const Status = () => {
  const [systemStatus, setSystemStatus] = useState(null);
  const [regions, setRegions] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStatus();
    // Refresh every 30 seconds
    const interval = setInterval(loadStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  const loadStatus = async () => {
    try {
      setLoading(true);
      const [status, regionData] = await Promise.all([
        getSystemStatus(),
        getRegionAvailability()
      ]);
      setSystemStatus(status);
      setRegions(regionData.regions || regionData || []);
    } catch (error) {
      console.error('Failed to load status:', error);
      toast.error('Failed to load system status');
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'Up':
      case 'Operational':
      case 'Healthy':
        return <CheckCircle2 className="h-5 w-5 text-green-600" />;
      case 'Degraded':
      case 'Warning':
        return <AlertCircle className="h-5 w-5 text-yellow-600" />;
      case 'Down':
      case 'Critical':
        return <XCircle className="h-5 w-5 text-red-600" />;
      default:
        return <Activity className="h-5 w-5 text-gray-600" />;
    }
  };

  const getStatusBadge = (status) => {
    switch (status) {
      case 'Up':
      case 'Operational':
      case 'Healthy':
        return <Badge className="bg-green-600">Operational</Badge>;
      case 'Degraded':
      case 'Warning':
        return <Badge className="bg-yellow-600">Degraded</Badge>;
      case 'Down':
      case 'Critical':
        return <Badge variant="destructive">Down</Badge>;
      default:
        return <Badge variant="outline">Unknown</Badge>;
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-32" />
        <div className="grid gap-4 md:grid-cols-3">
          {[1, 2, 3].map(i => <Skeleton key={i} className="h-48" />)}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">System Status</h1>
          <p className="text-muted-foreground">
            Monitor service health and availability
          </p>
        </div>
        {systemStatus && getStatusBadge(systemStatus.status)}
      </div>

      {/* Overall Status */}
      <Card className="border-2">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-2xl">All Systems Operational</CardTitle>
              <CardDescription>
                Last checked: {systemStatus ? new Date(systemStatus.last_checked).toLocaleString() : 'Unknown'}
              </CardDescription>
            </div>
            {getStatusIcon(systemStatus?.status)}
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div>
              <div className="text-sm text-muted-foreground mb-1">Uptime</div>
              <div className="text-xl font-semibold">{systemStatus?.uptime || 'N/A'}</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground mb-1">Version</div>
              <div className="text-xl font-semibold">{systemStatus?.version || '1.0.0'}</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground mb-1">Active Projects</div>
              <div className="text-xl font-semibold">{systemStatus?.active_projects || 0}</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Service Status */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Database</CardTitle>
              {getStatusIcon(systemStatus?.database?.status)}
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Status</span>
                <span className="font-medium">{systemStatus?.database?.status || 'Unknown'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Response Time</span>
                <span className="font-medium">{systemStatus?.database?.response_time_ms || 0}ms</span>
              </div>
              <div className="text-xs text-muted-foreground mt-4">
                {systemStatus?.database?.message || 'No issues detected'}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>API</CardTitle>
              {getStatusIcon(systemStatus?.api?.status)}
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Status</span>
                <span className="font-medium">{systemStatus?.api?.status || 'Unknown'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Response Time</span>
                <span className="font-medium">{systemStatus?.api?.response_time_ms || 0}ms</span>
              </div>
              <div className="text-xs text-muted-foreground mt-4">
                {systemStatus?.api?.message || 'No issues detected'}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>LiveKit</CardTitle>
              {getStatusIcon(systemStatus?.livekit?.status)}
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Status</span>
                <span className="font-medium">{systemStatus?.livekit?.status || 'Unknown'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Response Time</span>
                <span className="font-medium">{systemStatus?.livekit?.response_time_ms || 0}ms</span>
              </div>
              <div className="text-xs text-muted-foreground mt-4">
                {systemStatus?.livekit?.message || 'No issues detected'}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Region Status */}
      <Card>
        <CardHeader>
          <CardTitle>Region Availability</CardTitle>
          <CardDescription>Status of all regional data centers</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {regions.length > 0 ? (
              regions.map((region) => (
                <div key={region.region} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center gap-4">
                    {getStatusIcon(region.status)}
                    <div>
                      <div className="font-semibold">{region.region}</div>
                      <div className="text-sm text-muted-foreground">{region.message}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium">{region.latency_ms}ms</div>
                    <div className="text-xs text-muted-foreground">
                      {region.active_rooms || 0} active rooms
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                No region data available
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default Status;
