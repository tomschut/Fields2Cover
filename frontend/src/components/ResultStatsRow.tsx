interface ResultStatsRowProps {
  result: { path: { length_m: number; task_time_s: number } } | null
}

export function ResultStatsRow({ result }: ResultStatsRowProps) {
  if (result === null) return null
  return (
    <div className="result-stats">
      Path: {result.path.length_m.toFixed(0)} m · {(result.path.task_time_s / 60).toFixed(1)} min
    </div>
  )
}
