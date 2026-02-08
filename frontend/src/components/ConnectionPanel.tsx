import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { toast } from 'sonner'
import { QBitService } from '../../bindings/qbittorrent-file-matcher'

interface ConnectionPanelProps {
  onConnect: () => void
}

export function ConnectionPanel({ onConnect }: ConnectionPanelProps) {
  const [url, setUrl] = useState('http://localhost:8080')
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [isConnecting, setIsConnecting] = useState(false)

  const handleConnect = async () => {
    if (!url || !username) {
      toast.error('Please fill in URL and username')
      return
    }

    setIsConnecting(true)
    try {
      await QBitService.Connect({ url, username, password })
      const version = await QBitService.GetVersion()
      toast.success(`Connected to qBittorrent ${version}`)
      onConnect()
    } catch (error) {
      toast.error(`Connection failed: ${error}`)
    } finally {
      setIsConnecting(false)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-[calc(100vh-4rem)]">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle>Connect to qBittorrent</CardTitle>
          <CardDescription>
            Enter your WebUI credentials to get started
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Server URL</label>
            <Input
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="http://localhost:8080"
            />
          </div>
          
          <div className="space-y-2">
            <label className="text-sm font-medium">Username</label>
            <Input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="admin"
            />
          </div>
          
          <div className="space-y-2">
            <label className="text-sm font-medium">Password</label>
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              onKeyDown={(e) => e.key === 'Enter' && handleConnect()}
            />
          </div>

          <Button 
            onClick={handleConnect} 
            className="w-full"
            disabled={isConnecting}
          >
            {isConnecting ? (
              <>
                <Spinner className="mr-2" />
                Connecting...
              </>
            ) : 'Connect'}
          </Button>

          <p className="text-center text-xs text-muted-foreground">
            Enable WebUI in qBittorrent under Tools → Options → Web UI
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
