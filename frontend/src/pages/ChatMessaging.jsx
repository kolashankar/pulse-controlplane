import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { MessageSquare, Settings, BarChart3, BookOpen } from 'lucide-react';

const ChatMessaging = () => {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Chat Messaging</h1>
          <p className="text-muted-foreground">
            Configure real-time chat and messaging features
          </p>
        </div>
        <Button>
          <BookOpen className="mr-2 h-4 w-4" />
          View Documentation
        </Button>
      </div>

      {/* Enable/Disable */}
      <Card className="border-blue-200 bg-blue-50/50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Chat Messaging</CardTitle>
              <CardDescription>Enable real-time chat for your applications</CardDescription>
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
              <CardTitle>Chat Settings</CardTitle>
            </div>
            <CardDescription>Configure chat behavior and limits</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <Label htmlFor="typing">Typing Indicators</Label>
              <Switch id="typing" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="read">Read Receipts</Label>
              <Switch id="read" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="reactions">Message Reactions</Label>
              <Switch id="reactions" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="threads">Message Threading</Label>
              <Switch id="threads" />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="file">File Attachments</Label>
              <Switch id="file" defaultChecked />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              <CardTitle>Usage Metrics</CardTitle>
            </div>
            <CardDescription>Chat usage statistics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="text-2xl font-bold">12,450</div>
                <div className="text-sm text-muted-foreground">Total Messages</div>
              </div>
              <div>
                <div className="text-2xl font-bold">348</div>
                <div className="text-sm text-muted-foreground">Active Channels</div>
              </div>
              <div>
                <div className="text-2xl font-bold">1,234</div>
                <div className="text-sm text-muted-foreground">Active Users</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Features */}
      <Card>
        <CardHeader>
          <CardTitle>Features</CardTitle>
          <CardDescription>Available chat features and capabilities</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {[
              'Direct Messages',
              'Group Channels',
              'Public Channels',
              'Message Search',
              'Link Previews',
              'Emoji Support',
              'Message Editing',
              'Message Deletion',
              'User Mentions',
              'Channel Mentions',
              'Push Notifications',
              'Webhooks'
            ].map((feature) => (
              <div key={feature} className="flex items-center gap-2 p-3 border rounded-lg">
                <MessageSquare className="h-4 w-4 text-blue-600" />
                <span className="text-sm font-medium">{feature}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default ChatMessaging;
