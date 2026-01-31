class IncidentsByNameChart < BaseChart
  set_chart_name :incidents_by_name_chart
  set_human_name 'Incidents by name (top 10)'
  set_kind :pie_chart

  def chart
    Incident.top(:name, 10)
  end
end
