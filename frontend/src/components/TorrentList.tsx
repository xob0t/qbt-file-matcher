import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Spinner } from '@/components/ui/spinner'
import {
  Item,
  ItemContent,
  ItemTitle,
  ItemDescription,
  ItemActions,
  ItemGroup,
} from '@/components/ui/item'
import { toast } from 'sonner'
import { QBitService } from '../../bindings/qbittorrent-file-matcher/backend'
import { formatSize, getErrorMessage } from '@/lib/utils'
import type { TorrentInfo } from '../App'

interface TorrentListProps {
  onSelectTorrent: (torrent: TorrentInfo) => void
}

function getStateBadge(state: string): { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' } {
  const states: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
    downloading: { label: 'Downloading', variant: 'default' },
    seeding: { label: 'Seeding', variant: 'default' },
    pausedDL: { label: 'Paused', variant: 'secondary' },
    pausedUP: { label: 'Paused', variant: 'secondary' },
    stalledDL: { label: 'Stalled', variant: 'outline' },
    stalledUP: { label: 'Stalled', variant: 'outline' },
    error: { label: 'Error', variant: 'destructive' },
    missingFiles: { label: 'Missing', variant: 'destructive' },
    uploading: { label: 'Uploading', variant: 'default' },
    queuedDL: { label: 'Queued', variant: 'secondary' },
    queuedUP: { label: 'Queued', variant: 'secondary' },
    checkingDL: { label: 'Checking', variant: 'outline' },
    checkingUP: { label: 'Checking', variant: 'outline' },
    checkingResumeData: { label: 'Checking', variant: 'outline' },
  }
  return states[state] || { label: state, variant: 'secondary' }
}

export function TorrentList({ onSelectTorrent }: TorrentListProps) {
  const [torrents, setTorrents] = useState<TorrentInfo[]>([])
  const [filteredTorrents, setFilteredTorrents] = useState<TorrentInfo[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [isLoading, setIsLoading] = useState(true)

  const loadTorrents = useCallback(async () => {
    setIsLoading(true)
    try {
      const result = await QBitService.GetTorrents()
      setTorrents(result)
      setFilteredTorrents(result)
    } catch (error) {
      toast.error(`Failed to load torrents: ${getErrorMessage(error)}`)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    loadTorrents()
  }, [loadTorrents])

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

  return (
    <Card className="flex-1 flex flex-col min-h-0">
      <CardHeader className="shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Select Torrent</CardTitle>
            <CardDescription>
              {filteredTorrents.length} of {torrents.length} torrents
            </CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={loadTorrents} disabled={isLoading}>
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col min-h-0 gap-4">
        <Input
          placeholder="Search torrents..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="shrink-0"
        />

        {isLoading ? (
          <div className="flex-1 flex flex-col items-center justify-center gap-3">
            <Spinner className="size-6" />
            <p className="text-sm text-muted-foreground">Loading torrents...</p>
          </div>
        ) : filteredTorrents.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center gap-2">
            <p className="text-sm text-muted-foreground">
              {searchQuery ? 'No matching torrents found' : 'No torrents available'}
            </p>
          </div>
        ) : (
          <ScrollArea className="flex-1 min-h-0">
            <ItemGroup>
              {filteredTorrents.map((torrent) => {
                const stateInfo = getStateBadge(torrent.state)
                const progress = Math.round(torrent.progress * 100)
                
                return (
                  <Item
                    key={torrent.hash}
                    variant="outline"
                    className="cursor-pointer hover:bg-accent mb-2 bg-background"
                    onClick={() => onSelectTorrent(torrent)}
                  >
                    <ItemContent>
                      <ItemTitle className="truncate">{torrent.name}</ItemTitle>
                      <ItemDescription>
                        {formatSize(torrent.size)} • {progress}% complete
                      </ItemDescription>
                      <Progress value={progress} className="mt-2 h-1" />
                    </ItemContent>
                    <ItemActions>
                      <Badge variant={stateInfo.variant}>{stateInfo.label}</Badge>
                    </ItemActions>
                  </Item>
                )
              })}
            </ItemGroup>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
