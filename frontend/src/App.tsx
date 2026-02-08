import { useState, useCallback } from 'react'
import { ConnectionPanel } from './components/ConnectionPanel'
import { TorrentList } from './components/TorrentList'
import { MatchingPanel } from './components/MatchingPanel'
import { Toaster } from './components/ui/sonner'

export interface TorrentInfo {
  hash: string
  name: string
  size: number
  progress: number
  state: string
  savePath: string
  contentPath: string
}

function App() {
  const [isConnected, setIsConnected] = useState(false)
  const [selectedTorrent, setSelectedTorrent] = useState<TorrentInfo | null>(null)

  const handleConnect = useCallback(() => {
    setIsConnected(true)
  }, [])

  const handleDisconnect = useCallback(() => {
    setIsConnected(false)
    setSelectedTorrent(null)
  }, [])

  const handleSelectTorrent = useCallback((torrent: TorrentInfo) => {
    setSelectedTorrent(torrent)
  }, [])

  const handleBack = useCallback(() => {
    setSelectedTorrent(null)
  }, [])

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto py-6 px-4">
        <header className="mb-6">
          <h1 className="text-2xl font-bold text-foreground">qBittorrent File Matcher</h1>
          <p className="text-muted-foreground">Match torrent files with existing files on your disk</p>
        </header>

        {!isConnected ? (
          <ConnectionPanel onConnect={handleConnect} />
        ) : !selectedTorrent ? (
          <TorrentList 
            onSelectTorrent={handleSelectTorrent} 
            onDisconnect={handleDisconnect}
          />
        ) : (
          <MatchingPanel 
            torrent={selectedTorrent} 
            onBack={handleBack}
          />
        )}
      </div>
      <Toaster />
    </div>
  )
}

export default App
