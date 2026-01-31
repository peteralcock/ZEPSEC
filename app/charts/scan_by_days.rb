class ScanByDays < BaseChart
  set_chart_name :scan_by_days
  set_human_name 'Scans by days (per month)'
  set_kind :line_chart

  def chart
    result = ScanResult.group_by_day(:job_start, range: 1.month.ago.midnight..Time.now).count
    [{name: 'Count', data: result}]
  end
end
