import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent } from '@/components/ui/card';
import { Eye, EyeOff, Copy, Check, RefreshCw } from 'lucide-react';
import { toast } from 'sonner';

const APIKeyDisplay = ({ apiKey, apiSecret, onRegenerate, showRegenerateButton = true }) => {
  const [showSecret, setShowSecret] = useState(false);
  const [copiedKey, setCopiedKey] = useState(false);
  const [copiedSecret, setCopiedSecret] = useState(false);

  const copyToClipboard = async (text, type) => {
    try {
      await navigator.clipboard.writeText(text);
      if (type === 'key') {
        setCopiedKey(true);
        setTimeout(() => setCopiedKey(false), 2000);
      } else {
        setCopiedSecret(true);
        setTimeout(() => setCopiedSecret(false), 2000);
      }
      toast.success(`${type === 'key' ? 'API Key' : 'API Secret'} copied to clipboard`);
    } catch (err) {
      toast.error('Failed to copy to clipboard');
    }
  };

  return (
    <Card>
      <CardContent className="pt-6 space-y-4">
        {/* API Key */}
        <div className="space-y-2">
          <Label>API Key</Label>
          <div className="flex gap-2">
            <Input
              value={apiKey}
              readOnly
              className="font-mono text-sm"
            />
            <Button
              variant="outline"
              size="icon"
              onClick={() => copyToClipboard(apiKey, 'key')}
            >
              {copiedKey ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            </Button>
          </div>
        </div>

        {/* API Secret */}
        {apiSecret && (
          <div className="space-y-2">
            <Label>API Secret</Label>
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Input
                  type={showSecret ? 'text' : 'password'}
                  value={apiSecret}
                  readOnly
                  className="font-mono text-sm pr-10"
                />
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-0 top-0 h-full"
                  onClick={() => setShowSecret(!showSecret)}
                >
                  {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
              <Button
                variant="outline"
                size="icon"
                onClick={() => copyToClipboard(apiSecret, 'secret')}
              >
                {copiedSecret ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              ⚠️ This secret will only be shown once. Store it securely.
            </p>
          </div>
        )}

        {/* Regenerate Button */}
        {showRegenerateButton && onRegenerate && (
          <div className="pt-2">
            <Button
              variant="destructive"
              size="sm"
              onClick={onRegenerate}
              className="w-full"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Regenerate API Keys
            </Button>
            <p className="text-xs text-muted-foreground mt-2">
              Warning: Regenerating keys will invalidate the current keys.
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default APIKeyDisplay;
