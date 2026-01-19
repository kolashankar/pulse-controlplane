import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Shield, Settings, BarChart3, BookOpen, AlertTriangle } from 'lucide-react';

const Moderation = () => {
  const mockLogs = [
    { id: 1, type: 'profanity', user: 'user@example.com', action: 'Message blocked', timestamp: new Date() },
    { id: 2, type: 'spam', user: 'spam@example.com', action: 'User warned', timestamp: new Date() },
    { id: 3, type: 'harassment', user: 'bad@example.com', action: 'User banned', timestamp: new Date() },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Moderation</h1>
          <p className="text-muted-foreground">
            Configure content moderation and safety features
          </p>
        </div>
        <Button>
          <BookOpen className="mr-2 h-4 w-4" />
          View Documentation
        </Button>
      </div>

      {/* Enable/Disable */}
      <Card className="border-red-200 bg-red-50/50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Moderation</CardTitle>
              <CardDescription>Enable automated content moderation and safety checks</CardDescription>
            </div>
            <Switch defaultChecked />
          </div>
        </CardHeader>
      </Card>

      {/* Settings */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Settings className="h-5 w-5" />
              <CardTitle>Moderation Rules</CardTitle>
            </div>
            <CardDescription>Configure automated moderation rules</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <Label htmlFor="profanity">Profanity Filter</Label>
              <Switch id="profanity" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="spam">Spam Detection</Label>
              <Switch id="spam" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="links">Block External Links</Label>
              <Switch id="links" />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="caps">Excessive Caps Lock</Label>
              <Switch id="caps" />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="flood">Message Flooding</Label>
              <Switch id="flood" defaultChecked />
            </div>
            <div className="space-y-2">
              <Label htmlFor="rate-limit">Rate Limit (messages/minute)</Label>
              <Input id="rate-limit" type="number" defaultValue="10" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              <CardTitle>Moderation Stats</CardTitle>
            </div>
            <CardDescription>Moderation activity statistics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="text-2xl font-bold">87</div>
                <div className="text-sm text-muted-foreground">Messages Blocked</div>
              </div>
              <div>
                <div className="text-2xl font-bold">23</div>
                <div className="text-sm text-muted-foreground">Users Warned</div>
              </div>
              <div>
                <div className="text-2xl font-bold">5</div>
                <div className="text-sm text-muted-foreground">Users Banned</div>
              </div>
              <div>
                <div className="text-2xl font-bold">98.5%</div>
                <div className="text-sm text-muted-foreground">Detection Accuracy</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Custom Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Custom Filters</CardTitle>
          <CardDescription>Add custom words or phrases to block</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="blocked-words">Blocked Words (comma-separated)</Label>
            <Textarea
              id="blocked-words"
              placeholder="word1, word2, word3"
              rows={3}
            />
          </div>
          <Button>Save Filters</Button>
        </CardContent>
      </Card>

      {/* Moderation Logs */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Moderation Actions</CardTitle>
          <CardDescription>Latest automated moderation events</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Type</TableHead>
                <TableHead>User</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Timestamp</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {mockLogs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell>
                    <Badge variant="destructive">
                      <AlertTriangle className="h-3 w-3 mr-1" />
                      {log.type}
                    </Badge>
                  </TableCell>
                  <TableCell>{log.user}</TableCell>
                  <TableCell>{log.action}</TableCell>
                  <TableCell>{log.timestamp.toLocaleString()}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Features */}
      <Card>
        <CardHeader>
          <CardTitle>Features</CardTitle>
          <CardDescription>Available moderation features</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {[
              'Profanity Filter',
              'Spam Detection',
              'Link Filtering',
              'Image Moderation',
              'Rate Limiting',
              'Auto-Ban',
              'Shadow Ban',
              'Manual Review Queue',
              'User Reputation',
              'Community Reports',
              'Keyword Filtering',
              'Sentiment Analysis'
            ].map((feature) => (
              <div key={feature} className="flex items-center gap-2 p-3 border rounded-lg">
                <Shield className="h-4 w-4 text-red-600" />
                <span className="text-sm font-medium">{feature}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default Moderation;
