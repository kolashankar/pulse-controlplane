import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Video, Settings, BarChart3, BookOpen } from 'lucide-react';

const VideoAudio = () => {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Video & Audio</h1>
          <p className="text-muted-foreground">
            Configure video calling and streaming features
          </p>
        </div>
        <Button>
          <BookOpen className="mr-2 h-4 w-4" />
          View Documentation
        </Button>
      </div>

      {/* Enable/Disable */}
      <Card className="border-purple-200 bg-purple-50/50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Video & Audio</CardTitle>
              <CardDescription>Enable real-time video and audio communication</CardDescription>
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
              <CardTitle>Room Settings</CardTitle>
            </div>
            <CardDescription>Configure video room behavior</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="layout">Default Layout</Label>
              <Select defaultValue="grid">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="grid">Grid View</SelectItem>
                  <SelectItem value="speaker">Speaker View</SelectItem>
                  <SelectItem value="single">Single View</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="quality">Video Quality</Label>
              <Select defaultValue="hd">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sd">SD (480p)</SelectItem>
                  <SelectItem value="hd">HD (720p)</SelectItem>
                  <SelectItem value="fhd">Full HD (1080p)</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="max-participants">Max Participants</Label>
              <Input id="max-participants" type="number" defaultValue="50" />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="recording">Enable Recording</Label>
              <Switch id="recording" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="screenshare">Screen Sharing</Label>
              <Switch id="screenshare" defaultChecked />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              <CardTitle>Streaming Analytics</CardTitle>
            </div>
            <CardDescription>Video and audio usage statistics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="text-2xl font-bold">1,250</div>
                <div className="text-sm text-muted-foreground">Total Participant Minutes</div>
              </div>
              <div>
                <div className="text-2xl font-bold">48</div>
                <div className="text-sm text-muted-foreground">Active Rooms</div>
              </div>
              <div>
                <div className="text-2xl font-bold">320</div>
                <div className="text-sm text-muted-foreground">Egress Minutes</div>
              </div>
              <div>
                <div className="text-2xl font-bold">12 GB</div>
                <div className="text-sm text-muted-foreground">Recordings Storage</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Egress Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Egress Configuration</CardTitle>
          <CardDescription>Configure HLS streaming and recording output</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <Label htmlFor="hls">HLS Streaming</Label>
            <Switch id="hls" defaultChecked />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="rtmp">RTMP Output</Label>
            <Switch id="rtmp" />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="cloud">Cloud Recording</Label>
            <Switch id="cloud" defaultChecked />
          </div>
          <div className="space-y-2">
            <Label htmlFor="egress-layout">Egress Layout</Label>
            <Select defaultValue="speaker">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="speaker">Speaker View</SelectItem>
                <SelectItem value="grid">Grid View</SelectItem>
                <SelectItem value="single">Single Participant</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Features */}
      <Card>
        <CardHeader>
          <CardTitle>Features</CardTitle>
          <CardDescription>Available video and audio features</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {[
              'WebRTC Rooms',
              'Screen Sharing',
              'Audio Only Mode',
              'Picture-in-Picture',
              'Virtual Backgrounds',
              'Noise Cancellation',
              'HLS Streaming',
              'RTMP Output',
              'Cloud Recording',
              'Local Recording',
              'Simulcast',
              'Adaptive Bitrate'
            ].map((feature) => (
              <div key={feature} className="flex items-center gap-2 p-3 border rounded-lg">
                <Video className="h-4 w-4 text-purple-600" />
                <span className="text-sm font-medium">{feature}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default VideoAudio;
