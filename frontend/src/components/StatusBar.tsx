import { Button } from '@/components/ui/button'
import { QBitService } from '../../bindings/qbittorrent-file-matcher'
import type { ConnectionInfo } from '../App'

interface StatusBarProps {
  connectionInfo: ConnectionInfo
  onDisconnect: () => void
}

export function StatusBar({ connectionInfo, onDisconnect }: StatusBarProps) {
  const handleDisconnect = async () => {
    try {
      await QBitService.Disconnect()
    } catch {
      // Ignore errors on disconnect
    }
    onDisconnect()
  }

  // Extract host from URL for cleaner display
  const getHost = (url: string) => {
    try {
      const parsed = new URL(url)
      return parsed.host
    } catch {
      return url
    }
  }

  return (
    <div className="shrink-0 border-t bg-muted/30 px-4 py-2 flex items-center justify-between text-sm">
      <div className="flex items-center gap-2 text-muted-foreground">
        <span className="inline-block w-2 h-2 rounded-full bg-success animate-pulse" />
        <span>
          Connected to <span className="text-foreground font-medium">{getHost(connectionInfo.url)}</span>
          {' '}as <span className="text-foreground font-medium">{connectionInfo.username}</span>
        </span>
        <span className="text-muted-foreground/60">•</span>
        <span className="text-muted-foreground/80">qBittorrent {connectionInfo.version}</span>
      </div>
      <Button 
        variant="ghost" 
        size="sm" 
        onClick={handleDisconnect}
        className="h-7 text-xs text-muted-foreground hover:text-foreground"
      >
        Disconnect
      </Button>
    </div>
  )
}
