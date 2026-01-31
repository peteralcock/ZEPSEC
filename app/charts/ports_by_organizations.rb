class PortsByOrganizations < BaseChart
  set_chart_name :open_ports_by_organizations
  set_human_name 'Open ports by organizations (top 10 organizations)'
  set_kind :column_chart

  def chart
    scope = ScanResult.where(state: :open)
    scope = ScanResultsQuery.new(scope)
      .last_results
      .joins('LEFT OUTER JOIN hosts ON hosts.ip = scan_results.ip')
      .joins('LEFT OUTER JOIN organizations ON organizations.id = hosts.organization_id')
    result = scope.top('organizations.name', 10)
    [{name: 'Count', data: result}]
  end
end
