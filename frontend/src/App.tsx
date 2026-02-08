import { useState, useCallback } from 'react'
import { ConnectionPanel } from './components/ConnectionPanel'
import { TorrentList } from './components/TorrentList'
import { MatchingPanel } from './components/MatchingPanel'
import { StatusBar } from './components/StatusBar'
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

export interface ConnectionInfo {
  url: string
  username: string
  version: string
}

function App() {
  const [isConnected, setIsConnected] = useState(false)
  const [connectionInfo, setConnectionInfo] = useState<ConnectionInfo | null>(null)
  const [selectedTorrent, setSelectedTorrent] = useState<TorrentInfo | null>(null)

  const handleConnect = useCallback((info: ConnectionInfo) => {
    setConnectionInfo(info)
    setIsConnected(true)
  }, [])

  const handleDisconnect = useCallback(() => {
    setIsConnected(false)
    setConnectionInfo(null)
    setSelectedTorrent(null)
  }, [])

  const handleSelectTorrent = useCallback((torrent: TorrentInfo) => {
    setSelectedTorrent(torrent)
  }, [])

  const handleBack = useCallback(() => {
    setSelectedTorrent(null)
  }, [])

  return (
    <div className="h-screen flex flex-col bg-background">
      {!isConnected ? (
        <ConnectionPanel onConnect={handleConnect} />
      ) : (
        <>
          <div className="flex-1 flex flex-col min-h-0 p-4">
            {!selectedTorrent ? (
            <TorrentList onSelectTorrent={handleSelectTorrent} />
            ) : (
              <MatchingPanel 
                torrent={selectedTorrent} 
                onBack={handleBack}
              />
            )}
          </div>
          {connectionInfo && (
            <StatusBar connectionInfo={connectionInfo} onDisconnect={handleDisconnect} />
          )}
        </>
      )}
      <Toaster position="bottom-right" />
    </div>
  )
}

export default App
