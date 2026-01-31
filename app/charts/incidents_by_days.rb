class IncidentsByDays < BaseChart
  set_chart_name :incidents_by_days
  set_human_name 'Incidents by days (per month)'
  set_kind :line_chart

  def chart
    result = Incident.group_by_day(:created_at, range: 1.month.ago.midnight..Time.now).count
    [{name: 'Count', data: result}]
  end
end
