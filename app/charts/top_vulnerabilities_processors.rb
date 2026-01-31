class TopVulnerabilitiesProcessors < BaseChart
  set_chart_name :top_vulnerabilities_processors
  set_human_name 'Vulnerability processors (top 10)'
  set_kind :column_chart

  def chart
    scope = Pundit.policy_scope(current_user, Vulnerability)
    result = scope.joins(:processor)
                  .top('users.name', 10)
    [{name: 'Count', data: result}]
  end
end
