class IncidentsByOrganizations < BaseChart
  set_chart_name :incidents_by_organizations
  set_human_name 'Incidents by organizations (top 10 incident-related)'
  set_kind :column_chart

  def chart
    #result = Incident.joins(:links).group('links.first_record_id').count.limit 2
    result = Incident.joins(:incident_organizations)
      .top('organizations.name', 10)
    [{name: 'Count', data: result}]
  end
end
