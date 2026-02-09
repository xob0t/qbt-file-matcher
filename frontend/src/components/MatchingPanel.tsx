import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Spinner } from '@/components/ui/spinner'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Item,
  ItemContent,
  ItemTitle,
  ItemDescription,
  ItemActions,
  ItemGroup,
} from '@/components/ui/item'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { toast } from 'sonner'
import { Dialogs } from '@wailsio/runtime'
import { QBitService, MatcherService } from '../../bindings/qbt-file-matcher/backend'
import type { TorrentFile, DiskFile, MatchInfo } from '../../bindings/qbt-file-matcher/backend/models'
import { formatSize, getErrorMessage } from '@/lib/utils'
import type { TorrentInfo } from '../App'

interface MatchingPanelProps {
  torrent: TorrentInfo
  onBack: () => void
}

export function MatchingPanel({ torrent, onBack }: MatchingPanelProps) {
  const [searchPath, setSearchPath] = useState(torrent.savePath)
  const [torrentFiles, setTorrentFiles] = useState<TorrentFile[]>([])
  const [matches, setMatches] = useState<MatchInfo[]>([])
  const [unmatched, setUnmatched] = useState<{ index: number; name: string; size: number }[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isScanning, setIsScanning] = useState(false)
  const [isApplying, setIsApplying] = useState(false)
  const [isSkipping, setIsSkipping] = useState(false)
  const [isRechecking, setIsRechecking] = useState(false)
  const [showRecheckButton, setShowRecheckButton] = useState(false)
  const [requireSameExtension, setRequireSameExtension] = useState(true)
  const [selectDialogOpen, setSelectDialogOpen] = useState(false)
  const [currentMatchIndex, setCurrentMatchIndex] = useState<number | null>(null)

  const loadTorrentFiles = useCallback(async () => {
    setIsLoading(true)
    try {
      const files = await QBitService.GetTorrentFiles(torrent.hash)
      setTorrentFiles(files)
    } catch (error) {
      toast.error(`Failed to load files: ${getErrorMessage(error)}`)
    } finally {
      setIsLoading(false)
    }
  }, [torrent.hash])

  useEffect(() => {
    loadTorrentFiles()
  }, [loadTorrentFiles])

  const handleScan = async () => {
    if (!searchPath) {
      toast.error('Please enter a directory path')
      return
    }

    setIsScanning(true)
    try {
      const exists = await MatcherService.DirExists(searchPath)
      if (!exists) {
        toast.error('Directory not found')
        setIsScanning(false)
        return
      }

      const diskFiles = await MatcherService.ScanDir(searchPath)
      toast.info(`Found ${diskFiles.length} files on disk`)

      const torrentFileInfos = torrentFiles.map(f => ({
        index: f.index,
        name: f.name,
        size: f.size,
      }))

      const result = await MatcherService.FindMatches({
        torrentFiles: torrentFileInfos,
        diskFiles: diskFiles,
        requireSameExtension: requireSameExtension,
      })

      setMatches(result.matches)
      setUnmatched(result.unmatched)

      if (result.matchedCount > 0) {
        toast.success(`Matched ${result.matchedCount} of ${result.totalFiles} files`)
      } else {
        toast.warning('No automatic matches found')
      }
    } catch (error) {
      toast.error(`Scan failed: ${getErrorMessage(error)}`)
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

  const handleClearSelection = (matchIndex: number, e: React.MouseEvent) => {
    e.stopPropagation()
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
      toast.error('No files selected')
      return
    }

    setIsApplying(true)
    let successCount = 0
    let errorCount = 0

    try {
      const renames = await MatcherService.GenRenames({
        matches: matchesWithSelection,
        searchPath: searchPath,
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
        toast.success(`Renamed ${successCount} files successfully`)
        setShowRecheckButton(true)
      }
      if (errorCount > 0) {
        toast.error(`Failed to rename ${errorCount} files`)
      }

      await loadTorrentFiles()
      setMatches([])
      setUnmatched([])
    } catch (error) {
      toast.error(`Apply failed: ${getErrorMessage(error)}`)
    } finally {
      setIsApplying(false)
    }
  }

  const handleSkipUnmatched = async () => {
    if (unmatched.length === 0) {
      toast.error('No unmatched files to skip')
      return
    }

    setIsSkipping(true)
    try {
      // Get the file indices of all unmatched files
      const fileIDs = unmatched.map(f => f.index).join(',')
      
      // Set priority to 0 (do not download)
      await QBitService.SetFilePriority(torrent.hash, fileIDs, 0)
      
      toast.success(`Skipped ${unmatched.length} unmatched file${unmatched.length !== 1 ? 's' : ''}`)
      
      await loadTorrentFiles()
      setUnmatched([])
    } catch (error) {
      toast.error(`Failed to skip files: ${getErrorMessage(error)}`)
    } finally {
      setIsSkipping(false)
    }
  }

  const handleRecheck = async () => {
    setIsRechecking(true)
    try {
      await QBitService.RecheckTorrent(torrent.hash)
      toast.success('Recheck started - qBittorrent will verify file integrity')
      setShowRecheckButton(false)
    } catch (error) {
      toast.error(`Failed to start recheck: ${getErrorMessage(error)}`)
    } finally {
      setIsRechecking(false)
    }
  }

  const selectedCount = matches.filter(m => m.selected !== null).length
  const hasResults = matches.length > 0 || unmatched.length > 0
  
  // Count how many selected matches would actually result in a rename
  // (i.e., the relative path of the disk file differs from the torrent file name)
  const pendingRenamesCount = matches.filter(m => {
    if (!m.selected) return false
    // Compute relative path from searchPath to disk file
    // The disk file path starts with searchPath, so we need to extract the relative part
    const diskPath = m.selected.path
    const searchPathNormalized = searchPath.replace(/\\/g, '/')
    const diskPathNormalized = diskPath.replace(/\\/g, '/')
    
    // Get relative path
    let relativePath = diskPathNormalized
    if (diskPathNormalized.toLowerCase().startsWith(searchPathNormalized.toLowerCase())) {
      relativePath = diskPathNormalized.slice(searchPathNormalized.length)
      if (relativePath.startsWith('/')) relativePath = relativePath.slice(1)
    }
    
    // Compare with torrent file name
    return relativePath !== m.torrentFile.name
  }).length

  return (
    <>
      <Card className="flex-1 flex flex-col min-h-0">
        <CardHeader className="shrink-0">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <Button variant="ghost" size="sm" onClick={onBack} className="mb-2 -ml-2">
                ← Back
              </Button>
              <CardTitle className="truncate">{torrent.name}</CardTitle>
              <CardDescription>
                {torrentFiles.length} files • {formatSize(torrent.size)}
              </CardDescription>
            </div>
            <div className="flex gap-2 shrink-0">
              {showRecheckButton && (
                <TooltipProvider disableHoverableContent>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button 
                        onClick={handleRecheck} 
                        disabled={isRechecking} 
                        variant="secondary"
                      >
                        {isRechecking ? (
                          <>
                            <Spinner className="mr-2" />
                            Rechecking...
                          </>
                        ) : (
                          'Recheck Torrent'
                        )}
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Verify file integrity after renaming files</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
              {unmatched.length > 0 && (
                <TooltipProvider disableHoverableContent>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button 
                        onClick={handleSkipUnmatched} 
                        disabled={isSkipping || isApplying} 
                        variant="outline"
                      >
                        {isSkipping ? (
                          <>
                            <Spinner className="mr-2" />
                            Skipping...
                          </>
                        ) : (
                          `Skip ${unmatched.length} Unmatched`
                        )}
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Set priority to "Do not download" for files with no match on disk</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
              {pendingRenamesCount > 0 && (
                <Button onClick={handleApplyRenames} disabled={isApplying || isSkipping}>
                  {isApplying ? (
                    <>
                      <Spinner className="mr-2" />
                      Applying...
                    </>
                  ) : (
                    `Apply ${pendingRenamesCount} Rename${pendingRenamesCount !== 1 ? 's' : ''}`
                  )}
                </Button>
              )}
            </div>
          </div>
        </CardHeader>

        <CardContent className="flex-1 flex flex-col min-h-0 gap-4">
          {/* Search controls */}
          <div className="shrink-0 space-y-3">
            <div className="flex gap-2">
              <Input
                value={searchPath}
                onChange={(e) => setSearchPath(e.target.value)}
                placeholder="Enter directory path to scan..."
                className="flex-1"
              />
              <Button 
                onClick={async () => {
                  try {
                    // Only set Directory if it exists, otherwise let dialog use default
                    const dirExists = searchPath ? await MatcherService.DirExists(searchPath) : false
                    const path = await Dialogs.OpenFile({
                      CanChooseDirectories: true,
                      CanChooseFiles: false,
                      Title: 'Select Directory',
                      Directory: dirExists ? searchPath : undefined,
                    })
                    if (path) setSearchPath(path as string)
                  } catch {
                    // User cancelled
                  }
                }} 
                variant="secondary"
                disabled={isScanning || isLoading}
              >
                Browse
              </Button>
              <Button onClick={handleScan} disabled={isScanning || isLoading} variant="secondary">
                {isScanning ? <Spinner /> : 'Scan'}
              </Button>
            </div>

            <div className="flex items-center gap-2">
              <Checkbox
                id="requireExt"
                checked={requireSameExtension}
                onCheckedChange={(checked) => setRequireSameExtension(checked === true)}
              />
              <label htmlFor="requireExt" className="text-sm text-muted-foreground cursor-pointer">
                Require same file extension
              </label>
            </div>

            <p className="text-xs text-muted-foreground">
              Scan the download directory (not content directory) where your files are located. 
              Files will be matched by size and renamed to their relative path from this directory.
            </p>
          </div>

          {/* Content area */}
          {isLoading ? (
            <div className="flex-1 flex flex-col items-center justify-center gap-3">
              <Spinner className="size-6" />
              <p className="text-sm text-muted-foreground">Loading torrent files...</p>
            </div>
          ) : !hasResults ? (
            /* Initial state - show torrent files */
            <div className="flex-1 flex flex-col min-h-0 gap-3">
              <div className="shrink-0 flex items-center justify-between">
                <p className="text-sm font-medium">Torrent Files</p>
                <p className="text-xs text-muted-foreground">{torrentFiles.length} files</p>
              </div>
              <ScrollArea className="flex-1 min-h-0">
                <ItemGroup>
                  {torrentFiles.map((file) => (
                    <Item key={file.index} variant="muted" size="sm" className="mb-1">
                      <ItemContent>
                        <ItemTitle className="truncate text-sm">{file.name}</ItemTitle>
                        <ItemDescription>{formatSize(file.size)}</ItemDescription>
                      </ItemContent>
                      <ItemActions className="gap-3">
                        <Progress value={file.progress * 100} className="w-16 h-1" />
                        <Badge variant={file.progress === 1 ? 'default' : 'secondary'}>
                          {Math.round(file.progress * 100)}%
                        </Badge>
                      </ItemActions>
                    </Item>
                  ))}
                </ItemGroup>
              </ScrollArea>
            </div>
          ) : (
            /* Match results */
            <div className="flex-1 flex flex-col min-h-0 gap-3">
              <div className="shrink-0 flex items-center justify-between">
                <p className="text-sm font-medium">Match Results</p>
                <p className="text-xs text-muted-foreground">
                  {selectedCount} of {matches.length + unmatched.length} matched
                </p>
              </div>

              <ScrollArea className="flex-1 min-h-0">
                <ItemGroup>
                  {/* Matched files */}
                  {matches.map((match, index) => (
                    <Item
                      key={match.torrentFile.index}
                      variant="outline"
                      size="sm"
                      className={`mb-2 cursor-pointer ${
                        match.selected 
                          ? 'border-success/50 bg-success-muted/30' 
                          : 'border-warning/50 bg-warning-muted/30'
                      }`}
                      onClick={() => handleSelectMatch(index)}
                    >
                      <ItemContent>
                        <ItemTitle className="truncate text-sm">{match.torrentFile.name}</ItemTitle>
                        <ItemDescription>
                          {formatSize(match.torrentFile.size)}
                          {match.selected && (
                            <span className="block mt-1 truncate">→ {match.selected.name}</span>
                          )}
                        </ItemDescription>
                      </ItemContent>
                      <ItemActions>
                        {match.selected ? (
                          <>
                            <Badge className="bg-success text-success-foreground">Matched</Badge>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={(e) => handleClearSelection(index, e)}
                              className="h-6 w-6 p-0"
                            >
                              ×
                            </Button>
                          </>
                        ) : (
                          <Badge variant="outline" className="border-warning text-warning">
                            {match.diskFiles.length} candidate{match.diskFiles.length !== 1 ? 's' : ''}
                          </Badge>
                        )}
                      </ItemActions>
                    </Item>
                  ))}

                  {/* Unmatched files */}
                  {unmatched.map((file) => (
                    <Item
                      key={file.index}
                      variant="outline"
                      size="sm"
                      className="mb-2 border-destructive/50 bg-destructive-muted/30"
                    >
                      <ItemContent>
                        <ItemTitle className="truncate text-sm text-muted-foreground">
                          {file.name}
                        </ItemTitle>
                        <ItemDescription>{formatSize(file.size)}</ItemDescription>
                      </ItemContent>
                      <ItemActions>
                        <Badge variant="destructive">No match</Badge>
                      </ItemActions>
                    </Item>
                  ))}
                </ItemGroup>
              </ScrollArea>
            </div>
          )}
        </CardContent>
      </Card>

      {/* File selection dialog */}
      <Dialog open={selectDialogOpen} onOpenChange={setSelectDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Select Matching File</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Choose the file on disk that matches this torrent file
          </p>
          <ScrollArea className="max-h-[350px]">
            <ItemGroup>
              {currentMatchIndex !== null && matches[currentMatchIndex]?.diskFiles.map((file, i) => (
                <Item
                  key={i}
                  variant="outline"
                  size="sm"
                  className="mb-2 cursor-pointer hover:bg-accent"
                  onClick={() => handleChooseFile(file)}
                >
                  <ItemContent>
                    <ItemTitle className="truncate text-sm">{file.name}</ItemTitle>
                    <ItemDescription className="truncate">{file.path}</ItemDescription>
                    <ItemDescription>{formatSize(file.size)}</ItemDescription>
                  </ItemContent>
                </Item>
              ))}
            </ItemGroup>
          </ScrollArea>
        </DialogContent>
      </Dialog>
    </>
  )
}
