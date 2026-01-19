import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { MapPin, ExternalLink } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const ProjectCard = ({ project }) => {
  const navigate = useNavigate();

  const getRegionLabel = (region) => {
    const regions = {
      'us-east': 'US East',
      'us-west': 'US West',
      'eu-west': 'EU West',
      'asia-south': 'Asia South'
    };
    return regions[region] || region;
  };

  return (
    <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate(`/projects/${project.id}`)}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-lg">{project.name}</CardTitle>
            <CardDescription className="mt-1 flex items-center gap-2">
              <MapPin className="h-3 w-3" />
              {getRegionLabel(project.region)}
            </CardDescription>
          </div>
          <Button variant="ghost" size="icon" onClick={(e) => {
            e.stopPropagation();
            navigate(`/projects/${project.id}`);
          }}>
            <ExternalLink className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          <div>
            <div className="text-xs text-muted-foreground mb-1">Project ID</div>
            <div className="text-sm font-mono bg-slate-100 px-2 py-1 rounded">
              {project.id}
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge variant="secondary">Chat</Badge>
            <Badge variant="secondary">Video</Badge>
            <Badge variant="secondary">Activity Feeds</Badge>
            <Badge variant="secondary">Moderation</Badge>
          </div>
          <div className="text-xs text-muted-foreground">
            Created {new Date(project.created_at).toLocaleDateString()}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default ProjectCard;
