import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'sonner'
import { 
  Loader2, 
  ArrowLeft, 
  FolderSearch, 
  Check, 
  X, 
  AlertCircle,
  FileQuestion,
  ChevronRight
} from 'lucide-react'
import { QBitService, MatcherService } from '../../bindings/qbittorrent-file-matcher'
import type { TorrentInfo } from '../App'

interface MatchingPanelProps {
  torrent: TorrentInfo
  onBack: () => void
}

interface TorrentFile {
  index: number
  name: string
  size: number
  progress: number
}

interface DiskFile {
  path: string
  name: string
  size: number
}

interface MatchInfo {
  torrentFile: { index: number; name: string; size: number }
  diskFiles: DiskFile[]
  selected: DiskFile | null
  autoMatched: boolean
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export function MatchingPanel({ torrent, onBack }: MatchingPanelProps) {
  const [searchPath, setSearchPath] = useState(torrent.contentPath || torrent.savePath)
  const [torrentFiles, setTorrentFiles] = useState<TorrentFile[]>([])
  const [matches, setMatches] = useState<MatchInfo[]>([])
  const [unmatched, setUnmatched] = useState<{ index: number; name: string; size: number }[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isScanning, setIsScanning] = useState(false)
  const [isApplying, setIsApplying] = useState(false)
  const [requireSameExtension, setRequireSameExtension] = useState(true)
  const [selectDialogOpen, setSelectDialogOpen] = useState(false)
  const [currentMatchIndex, setCurrentMatchIndex] = useState<number | null>(null)

  useEffect(() => {
    loadTorrentFiles()
  }, [torrent.hash])

  const loadTorrentFiles = async () => {
    setIsLoading(true)
    try {
      const files = await QBitService.GetTorrentFiles(torrent.hash)
      setTorrentFiles(files)
    } catch (error) {
      toast.error(`Failed to load torrent files: ${error}`)
    } finally {
      setIsLoading(false)
    }
  }

  const handleScan = async () => {
    if (!searchPath) {
      toast.error('Please enter a search path')
      return
    }

    setIsScanning(true)
    try {
      // Check if directory exists
      const exists = await MatcherService.DirectoryExists(searchPath)
      if (!exists) {
        toast.error('Directory does not exist')
        setIsScanning(false)
        return
      }

      // Scan directory
      const diskFiles = await MatcherService.ScanDirectory(searchPath)
      toast.info(`Found ${diskFiles.length} files on disk`)

      // Convert torrent files to the format expected by the matcher
      const torrentFileInfos = torrentFiles.map(f => ({
        index: f.index,
        name: f.name,
        size: f.size,
      }))

      // Find matches
      const result = await MatcherService.FindMatches({
        torrentFiles: torrentFileInfos,
        diskFiles: diskFiles,
        requireSameExtension: requireSameExtension,
      })

      setMatches(result.matches)
      setUnmatched(result.unmatched)

      if (result.matchedCount > 0) {
        toast.success(`Found ${result.matchedCount} automatic matches out of ${result.totalFiles} files`)
      } else {
        toast.warning('No automatic matches found')
      }
    } catch (error) {
      toast.error(`Scan failed: ${error}`)
    } finally {
      setIsScanning(false)
    }
  }

  const handleSelectMatch = (matchIndex: number) => {
    setCurrentMatchIndex(matchIndex)
    setSelectDialogOpen(true)
  }

  const handleChooseFile = (diskFile: DiskFile) => {
    if (currentMatchIndex === null) return

    setMatches(prev => {
      const updated = [...prev]
      updated[currentMatchIndex] = {
        ...updated[currentMatchIndex],
        selected: diskFile,
        autoMatched: false,
      }
      return updated
    })
    setSelectDialogOpen(false)
    setCurrentMatchIndex(null)
  }

  const handleClearSelection = (matchIndex: number) => {
    setMatches(prev => {
      const updated = [...prev]
      updated[matchIndex] = {
        ...updated[matchIndex],
        selected: null,
        autoMatched: false,
      }
      return updated
    })
  }

  const handleApplyRenames = async () => {
    const matchesWithSelection = matches.filter(m => m.selected !== null)
    if (matchesWithSelection.length === 0) {
      toast.error('No files selected for renaming')
      return
    }

    setIsApplying(true)
    let successCount = 0
    let errorCount = 0

    try {
      // Generate rename operations
      const renames = await MatcherService.GenerateRenames({
        matches: matchesWithSelection,
        torrentContentPath: torrent.contentPath || torrent.savePath,
      })

      for (const rename of renames) {
        try {
          await QBitService.RenameFile(torrent.hash, rename.oldPath, rename.newPath)
          successCount++
        } catch (error) {
          errorCount++
          console.error(`Failed to rename ${rename.oldPath}:`, error)
        }
      }

      if (successCount > 0) {
        toast.success(`Successfully renamed ${successCount} files`)
      }
      if (errorCount > 0) {
        toast.error(`Failed to rename ${errorCount} files`)
      }

      // Reload torrent files to reflect changes
      await loadTorrentFiles()
      setMatches([])
      setUnmatched([])
    } catch (error) {
      toast.error(`Apply failed: ${error}`)
    } finally {
      setIsApplying(false)
    }
  }

  const selectedCount = matches.filter(m => m.selected !== null).length

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={onBack}>
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <div>
              <CardTitle className="truncate">{torrent.name}</CardTitle>
              <CardDescription>
                Match torrent files with files on your disk
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Search Path Input */}
          <div className="flex gap-2">
            <div className="flex-1">
              <Input
                value={searchPath}
                onChange={(e) => setSearchPath(e.target.value)}
                placeholder="Enter directory path to search..."
              />
            </div>
            <Button onClick={handleScan} disabled={isScanning || isLoading}>
              {isScanning ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <FolderSearch className="mr-2 h-4 w-4" />
              )}
              Scan
            </Button>
          </div>

          {/* Options */}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="requireExt"
              checked={requireSameExtension}
              onChange={(e) => setRequireSameExtension(e.target.checked)}
              className="rounded border-input"
            />
            <label htmlFor="requireExt" className="text-sm">
              Require same file extension
            </label>
          </div>

          {/* Loading State */}
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : matches.length === 0 && unmatched.length === 0 ? (
            /* Initial State - Show Torrent Files */
            <div>
              <h3 className="font-medium mb-2">Torrent Files ({torrentFiles.length})</h3>
              <ScrollArea className="h-[400px] border rounded-lg">
                <div className="p-2 space-y-1">
                  {torrentFiles.map((file) => (
                    <div
                      key={file.index}
                      className="flex items-center justify-between p-2 rounded hover:bg-accent"
                    >
                      <div className="flex-1 min-w-0 mr-4">
                        <div className="text-sm truncate">{file.name}</div>
                        <div className="text-xs text-muted-foreground">
                          {formatSize(file.size)}
                        </div>
                      </div>
                      <Badge variant={file.progress === 1 ? 'default' : 'secondary'}>
                        {Math.round(file.progress * 100)}%
                      </Badge>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </div>
          ) : (
            /* Match Results */
            <div className="space-y-4">
              {/* Summary */}
              <div className="flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                  {selectedCount} of {matches.length + unmatched.length} files matched
                </div>
                {selectedCount > 0 && (
                  <Button onClick={handleApplyRenames} disabled={isApplying}>
                    {isApplying ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                      <Check className="mr-2 h-4 w-4" />
                    )}
                    Apply {selectedCount} Renames
                  </Button>
                )}
              </div>

              {/* Matches */}
              <ScrollArea className="h-[400px] border rounded-lg">
                <div className="p-2 space-y-2">
                  {matches.map((match, index) => (
                    <div
                      key={match.torrentFile.index}
                      className={`p-3 rounded-lg border ${
                        match.selected ? 'border-green-500/50 bg-green-500/5' : 'border-yellow-500/50 bg-yellow-500/5'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-medium truncate">
                            {match.torrentFile.name}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {formatSize(match.torrentFile.size)}
                          </div>
                        </div>
                        <div className="flex items-center gap-1">
                          {match.selected ? (
                            <>
                              <Badge variant="outline" className="text-green-500">
                                <Check className="mr-1 h-3 w-3" />
                                Matched
                              </Badge>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6"
                                onClick={() => handleClearSelection(index)}
                              >
                                <X className="h-3 w-3" />
                              </Button>
                            </>
                          ) : (
                            <Badge variant="outline" className="text-yellow-500">
                              <AlertCircle className="mr-1 h-3 w-3" />
                              {match.diskFiles.length} candidates
                            </Badge>
                          )}
                        </div>
                      </div>
                      
                      {match.selected ? (
                        <div 
                          className="mt-2 text-xs text-muted-foreground truncate cursor-pointer hover:text-foreground"
                          onClick={() => handleSelectMatch(index)}
                        >
                          <ChevronRight className="inline h-3 w-3 mr-1" />
                          {match.selected.path}
                        </div>
                      ) : (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="mt-2"
                          onClick={() => handleSelectMatch(index)}
                        >
                          Select a file...
                        </Button>
                      )}
                    </div>
                  ))}

                  {/* Unmatched Files */}
                  {unmatched.map((file) => (
                    <div
                      key={file.index}
                      className="p-3 rounded-lg border border-red-500/50 bg-red-500/5"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-medium truncate">
                            {file.name}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {formatSize(file.size)}
                          </div>
                        </div>
                        <Badge variant="outline" className="text-red-500">
                          <FileQuestion className="mr-1 h-3 w-3" />
                          No match
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </div>
          )}
        </CardContent>
      </Card>

      {/* File Selection Dialog */}
      <Dialog open={selectDialogOpen} onOpenChange={setSelectDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>Select a File</DialogTitle>
            <DialogDescription>
              Choose which file on disk matches the torrent file
            </DialogDescription>
          </DialogHeader>
          <ScrollArea className="max-h-[60vh]">
            <div className="space-y-2 p-1">
              {currentMatchIndex !== null && matches[currentMatchIndex]?.diskFiles.map((file, i) => (
                <div
                  key={i}
                  className="p-3 rounded-lg border hover:bg-accent cursor-pointer transition-colors"
                  onClick={() => handleChooseFile(file)}
                >
                  <div className="font-medium truncate">{file.name}</div>
                  <div className="text-xs text-muted-foreground truncate mt-1">
                    {file.path}
                  </div>
                  <div className="text-xs text-muted-foreground mt-1">
                    {formatSize(file.size)}
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        </DialogContent>
      </Dialog>
    </>
  )
}
