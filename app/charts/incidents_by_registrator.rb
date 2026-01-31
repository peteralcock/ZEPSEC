class IncidentsByRegistrator < BaseChart
  set_chart_name :incidents_by_registrator
  set_human_name 'Incidents by employees (top 10 registrators)'
  set_kind :bar_chart

  def chart
    result = Incident.joins(:user).top('users.name', 10)
    [{name: 'Count', data: result}]
  end
end
