interface PipelineErrorBannerProps {
  error: string | null
}

export function PipelineErrorBanner({ error }: PipelineErrorBannerProps) {
  if (error === null) return null
  return <div className="pipeline-error">{error}</div>
}
