import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'
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
      toast.error('Please fill in all required fields')
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
    <Card className="max-w-md mx-auto">
      <CardHeader>
        <CardTitle>Connect to qBittorrent</CardTitle>
        <CardDescription>
          Enter your qBittorrent Web UI credentials. Make sure Web UI is enabled in qBittorrent settings.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Web UI URL</label>
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
            placeholder="Enter password"
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
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Connecting...
            </>
          ) : (
            'Connect'
          )}
        </Button>
      </CardContent>
    </Card>
  )
}
