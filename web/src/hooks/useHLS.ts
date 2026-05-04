import Hls from 'hls.js'
import { useEffect, useRef } from 'react'

export function useHLS(src: string | null) {
  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    const video = videoRef.current
    if (!video || !src) return

    if (Hls.isSupported()) {
      const hls = new Hls({
        enableWorker: true,
        // Start playback at live edge
        startPosition: -1,
        // Stay 1 segment behind live
        liveSyncDurationCount: 1,
        // Jump forward if we drift more than 3 segments behind
        liveMaxLatencyDurationCount: 3,
        // Small buffer to reduce latency
        maxBufferLength: 4,
        maxMaxBufferLength: 8,
      })

      hls.loadSource(src)
      hls.attachMedia(video)

      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        video.play().catch(() => {})
      })

      return () => hls.destroy()
    }

    // Safari native HLS
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = src
      video.play().catch(() => {})
    }
  }, [src])

  return videoRef
}
