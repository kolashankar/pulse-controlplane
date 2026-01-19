import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { Separator } from '@/components/ui/separator';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog';
import { ArrowLeft, Save, Trash2 } from 'lucide-react';
import APIKeyDisplay from '@/components/APIKeyDisplay';
import { getProject, updateProject, deleteProject, regenerateAPIKeys } from '@/api/projects';
import { toast } from 'sonner';

const ProjectDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [project, setProject] = useState(null);
  const [loading, setLoading] = useState(true);
  const [formData, setFormData] = useState({
    name: '',
    region: '',
    webhook_url: '',
    storage_config: {
      provider: '',
      bucket: '',
      access_key_id: '',
      secret_access_key: ''
    }
  });

  useEffect(() => {
    if (id && id !== 'new') {
      loadProject();
    } else {
      setLoading(false);
    }
  }, [id]);

  const loadProject = async () => {
    try {
      setLoading(true);
      const data = await getProject(id);
      setProject(data.project || data);
      setFormData({
        name: data.project?.name || data.name || '',
        region: data.project?.region || data.region || '',
        webhook_url: data.project?.webhook_url || data.webhook_url || '',
        storage_config: data.project?.storage_config || data.storage_config || {
          provider: '',
          bucket: '',
          access_key_id: '',
          secret_access_key: ''
        }
      });
    } catch (error) {
      console.error('Failed to load project:', error);
      toast.error('Failed to load project');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdate = async () => {
    try {
      await updateProject(id, formData);
      toast.success('Project updated successfully');
      loadProject();
    } catch (error) {
      console.error('Failed to update project:', error);
      toast.error('Failed to update project');
    }
  };

  const handleRegenerate = async () => {
    try {
      const data = await regenerateAPIKeys(id);
      toast.success('API keys regenerated successfully');
      setProject({ ...project, ...data });
    } catch (error) {
      console.error('Failed to regenerate keys:', error);
      toast.error('Failed to regenerate API keys');
    }
  };

  const handleDelete = async () => {
    try {
      await deleteProject(id);
      toast.success('Project deleted successfully');
      navigate('/projects');
    } catch (error) {
      console.error('Failed to delete project:', error);
      toast.error('Failed to delete project');
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-96" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate('/projects')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{project?.name || 'New Project'}</h1>
            <p className="text-muted-foreground">
              {project ? `Project ID: ${project.id}` : 'Create a new project'}
            </p>
          </div>
        </div>
        {project && (
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive">
                <Trash2 className="mr-2 h-4 w-4" />
                Delete Project
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                <AlertDialogDescription>
                  This action cannot be undone. This will permanently delete the project
                  and all associated data.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction onClick={handleDelete}>Delete</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}
      </div>

      {/* Content */}
      <Tabs defaultValue="settings" className="space-y-4">
        <TabsList>
          <TabsTrigger value="settings">Settings</TabsTrigger>
          <TabsTrigger value="api-keys" disabled={!project}>API Keys</TabsTrigger>
          <TabsTrigger value="storage" disabled={!project}>Storage</TabsTrigger>
        </TabsList>

        <TabsContent value="settings" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Project Settings</CardTitle>
              <CardDescription>Configure your project settings</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Project Name</Label>
                <Input
                  id="name"
                  placeholder="My Project"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="region">Region</Label>
                <Select value={formData.region} onValueChange={(value) => setFormData({ ...formData, region: value })}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select region" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="us-east">US East</SelectItem>
                    <SelectItem value="us-west">US West</SelectItem>
                    <SelectItem value="eu-west">EU West</SelectItem>
                    <SelectItem value="asia-south">Asia South</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="webhook">Webhook URL</Label>
                <Input
                  id="webhook"
                  placeholder="https://example.com/webhook"
                  value={formData.webhook_url}
                  onChange={(e) => setFormData({ ...formData, webhook_url: e.target.value })}
                />
              </div>
              <Button onClick={handleUpdate} disabled={!project}>
                <Save className="mr-2 h-4 w-4" />
                Save Changes
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="api-keys" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>API Keys</CardTitle>
              <CardDescription>
                Use these keys to authenticate your API requests
              </CardDescription>
            </CardHeader>
            <CardContent>
              {project && (
                <APIKeyDisplay
                  apiKey={project.pulse_api_key}
                  apiSecret={project.pulse_api_secret}
                  onRegenerate={handleRegenerate}
                />
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="storage" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Storage Configuration</CardTitle>
              <CardDescription>
                Configure your cloud storage for recordings and media
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="provider">Storage Provider</Label>
                <Select
                  value={formData.storage_config.provider}
                  onValueChange={(value) => setFormData({
                    ...formData,
                    storage_config: { ...formData.storage_config, provider: value }
                  })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select provider" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="r2">Cloudflare R2</SelectItem>
                    <SelectItem value="s3">AWS S3</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="bucket">Bucket Name</Label>
                <Input
                  id="bucket"
                  placeholder="my-bucket"
                  value={formData.storage_config.bucket}
                  onChange={(e) => setFormData({
                    ...formData,
                    storage_config: { ...formData.storage_config, bucket: e.target.value }
                  })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="access-key">Access Key ID</Label>
                <Input
                  id="access-key"
                  type="password"
                  placeholder="Access Key ID"
                  value={formData.storage_config.access_key_id}
                  onChange={(e) => setFormData({
                    ...formData,
                    storage_config: { ...formData.storage_config, access_key_id: e.target.value }
                  })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="secret-key">Secret Access Key</Label>
                <Input
                  id="secret-key"
                  type="password"
                  placeholder="Secret Access Key"
                  value={formData.storage_config.secret_access_key}
                  onChange={(e) => setFormData({
                    ...formData,
                    storage_config: { ...formData.storage_config, secret_access_key: e.target.value }
                  })}
                />
              </div>
              <Button onClick={handleUpdate}>
                <Save className="mr-2 h-4 w-4" />
                Save Storage Config
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default ProjectDetails;
