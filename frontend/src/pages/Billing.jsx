import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Download, CreditCard } from 'lucide-react';
import UsageChart from '@/components/UsageChart';
import { getBillingDashboard, listInvoices } from '@/api/billing';
import { getUsageSummary } from '@/api/usage';
import { toast } from 'sonner';

const Billing = () => {
  const [loading, setLoading] = useState(true);
  const [billingData, setBillingData] = useState(null);
  const [invoices, setInvoices] = useState([]);
  const [usageData, setUsageData] = useState([]);

  // Mock project ID - in real app, get from context/params
  const mockProjectId = '507f1f77bcf86cd799439011';

  useEffect(() => {
    loadBillingData();
  }, []);

  const loadBillingData = async () => {
    try {
      setLoading(true);
      const [billing, invoiceList, usage] = await Promise.all([
        getBillingDashboard(mockProjectId).catch(() => null),
        listInvoices(mockProjectId).catch(() => ({ invoices: [] })),
        getUsageSummary(mockProjectId).catch(() => null)
      ]);

      setBillingData(billing);
      setInvoices(invoiceList.invoices || []);
      
      // Mock usage data for chart
      setUsageData([
        { date: '2025-01-01', value: 1200 },
        { date: '2025-01-02', value: 1500 },
        { date: '2025-01-03', value: 1800 },
        { date: '2025-01-04', value: 2200 },
        { date: '2025-01-05', value: 1900 },
        { date: '2025-01-06', value: 2400 },
        { date: '2025-01-07', value: 2100 }
      ]);
    } catch (error) {
      console.error('Failed to load billing data:', error);
      toast.error('Failed to load billing data');
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status) => {
    const variants = {
      paid: 'default',
      pending: 'secondary',
      overdue: 'destructive'
    };
    return <Badge variant={variants[status] || 'outline'}>{status}</Badge>;
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-64" />
        <div className="grid gap-4 md:grid-cols-3">
          {[1, 2, 3].map(i => <Skeleton key={i} className="h-32" />)}
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
          <h1 className="text-3xl font-bold tracking-tight">Billing & Usage</h1>
          <p className="text-muted-foreground">
            Monitor your usage and manage billing
          </p>
        </div>
        <Button variant="outline">
          <Download className="mr-2 h-4 w-4" />
          Download Invoice
        </Button>
      </div>

      {/* Current Plan */}
      <Card className="border-blue-200 bg-blue-50/50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Current Plan: Pro</CardTitle>
              <CardDescription>$49/month + usage</CardDescription>
            </div>
            <Button>Upgrade Plan</Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div>
              <div className="text-2xl font-bold">$127.50</div>
              <div className="text-sm text-muted-foreground">Current month charges</div>
            </div>
            <div>
              <div className="text-2xl font-bold">1,250</div>
              <div className="text-sm text-muted-foreground">Participant minutes</div>
            </div>
            <div>
              <div className="text-2xl font-bold">45 GB</div>
              <div className="text-sm text-muted-foreground">Bandwidth used</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Tabs */}
      <Tabs defaultValue="usage" className="space-y-4">
        <TabsList>
          <TabsTrigger value="usage">Usage</TabsTrigger>
          <TabsTrigger value="invoices">Invoices</TabsTrigger>
        </TabsList>

        <TabsContent value="usage" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <UsageChart
              data={usageData}
              type="line"
              title="Participant Minutes"
              description="Last 7 days"
              dataKey="value"
            />
            <UsageChart
              data={usageData}
              type="bar"
              title="API Requests"
              description="Last 7 days"
              dataKey="value"
            />
          </div>

          {/* Usage Breakdown */}
          <Card>
            <CardHeader>
              <CardTitle>Usage Breakdown</CardTitle>
              <CardDescription>Detailed usage metrics for current billing period</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between pb-4 border-b">
                  <div>
                    <div className="font-medium">Participant Minutes</div>
                    <div className="text-sm text-muted-foreground">1,250 / 100,000</div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">$5.00</div>
                    <div className="text-sm text-muted-foreground">$0.004 per minute</div>
                  </div>
                </div>
                <div className="flex items-center justify-between pb-4 border-b">
                  <div>
                    <div className="font-medium">Egress Minutes</div>
                    <div className="text-sm text-muted-foreground">320 / 10,000</div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">$3.84</div>
                    <div className="text-sm text-muted-foreground">$0.012 per minute</div>
                  </div>
                </div>
                <div className="flex items-center justify-between pb-4 border-b">
                  <div>
                    <div className="font-medium">Storage</div>
                    <div className="text-sm text-muted-foreground">12 GB / 100 GB</div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">$1.20</div>
                    <div className="text-sm text-muted-foreground">$0.10 per GB</div>
                  </div>
                </div>
                <div className="flex items-center justify-between pb-4 border-b">
                  <div>
                    <div className="font-medium">Bandwidth</div>
                    <div className="text-sm text-muted-foreground">45 GB / 1 TB</div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">$2.25</div>
                    <div className="text-sm text-muted-foreground">$0.05 per GB</div>
                  </div>
                </div>
                <div className="flex items-center justify-between pt-4">
                  <div className="font-semibold text-lg">Total Usage</div>
                  <div className="font-bold text-lg">$12.29</div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="invoices" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Invoice History</CardTitle>
              <CardDescription>View and download past invoices</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {invoices.length > 0 ? (
                  invoices.map((invoice) => (
                    <div key={invoice.id} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center gap-4">
                        <CreditCard className="h-8 w-8 text-muted-foreground" />
                        <div>
                          <div className="font-medium">Invoice #{invoice.id.slice(-8)}</div>
                          <div className="text-sm text-muted-foreground">
                            {new Date(invoice.created_at).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="text-right">
                          <div className="font-semibold">${invoice.amount?.toFixed(2) || '0.00'}</div>
                          {getStatusBadge(invoice.status)}
                        </div>
                        <Button variant="outline" size="sm">
                          <Download className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-12 text-muted-foreground">
                    <CreditCard className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p>No invoices yet</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default Billing;
