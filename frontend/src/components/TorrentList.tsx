import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { toast } from 'sonner'
import { Loader2, Search, LogOut, HardDrive } from 'lucide-react'
import { QBitService } from '../../bindings/qbittorrent-file-matcher'
import type { TorrentInfo } from '../App'

interface TorrentListProps {
  onSelectTorrent: (torrent: TorrentInfo) => void
  onDisconnect: () => void
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function getStateBadge(state: string) {
  const stateMap: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
    downloading: { label: 'Downloading', variant: 'default' },
    seeding: { label: 'Seeding', variant: 'default' },
    pausedDL: { label: 'Paused', variant: 'secondary' },
    pausedUP: { label: 'Paused', variant: 'secondary' },
    stalledDL: { label: 'Stalled', variant: 'outline' },
    stalledUP: { label: 'Stalled', variant: 'outline' },
    error: { label: 'Error', variant: 'destructive' },
    missingFiles: { label: 'Missing Files', variant: 'destructive' },
    uploading: { label: 'Uploading', variant: 'default' },
    queuedDL: { label: 'Queued', variant: 'secondary' },
    queuedUP: { label: 'Queued', variant: 'secondary' },
    checkingDL: { label: 'Checking', variant: 'outline' },
    checkingUP: { label: 'Checking', variant: 'outline' },
    checkingResumeData: { label: 'Checking', variant: 'outline' },
  }
  
  const info = stateMap[state] || { label: state, variant: 'outline' as const }
  return <Badge variant={info.variant}>{info.label}</Badge>
}

export function TorrentList({ onSelectTorrent, onDisconnect }: TorrentListProps) {
  const [torrents, setTorrents] = useState<TorrentInfo[]>([])
  const [filteredTorrents, setFilteredTorrents] = useState<TorrentInfo[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadTorrents()
  }, [])

  useEffect(() => {
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      setFilteredTorrents(
        torrents.filter(t => t.name.toLowerCase().includes(query))
      )
    } else {
      setFilteredTorrents(torrents)
    }
  }, [searchQuery, torrents])

  const loadTorrents = async () => {
    setIsLoading(true)
    try {
      const result = await QBitService.GetTorrents()
      setTorrents(result)
      setFilteredTorrents(result)
    } catch (error) {
      toast.error(`Failed to load torrents: ${error}`)
    } finally {
      setIsLoading(false)
    }
  }

  const handleDisconnect = async () => {
    try {
      await QBitService.Disconnect()
      onDisconnect()
    } catch (error) {
      toast.error(`Failed to disconnect: ${error}`)
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Select a Torrent</CardTitle>
            <CardDescription>
              Choose a torrent to match with files on your disk
            </CardDescription>
          </div>
          <Button variant="outline" onClick={handleDisconnect}>
            <LogOut className="mr-2 h-4 w-4" />
            Disconnect
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="mb-4 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search torrents..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : filteredTorrents.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            {searchQuery ? 'No torrents match your search' : 'No torrents found'}
          </div>
        ) : (
          <ScrollArea className="h-[500px]">
            <div className="space-y-2">
              {filteredTorrents.map((torrent) => (
                <div
                  key={torrent.hash}
                  className="flex items-center justify-between p-3 rounded-lg border hover:bg-accent cursor-pointer transition-colors"
                  onClick={() => onSelectTorrent(torrent)}
                >
                  <div className="flex-1 min-w-0 mr-4">
                    <div className="font-medium truncate">{torrent.name}</div>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground mt-1">
                      <HardDrive className="h-3 w-3" />
                      <span>{formatSize(torrent.size)}</span>
                      <span className="text-muted-foreground/50">|</span>
                      <span>{Math.round(torrent.progress * 100)}%</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {getStateBadge(torrent.state)}
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
