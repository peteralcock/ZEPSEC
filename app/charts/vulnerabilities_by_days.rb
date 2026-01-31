class VulnerabilitiesByDays < BaseChart
  set_chart_name :vulnerabilities_by_days
  set_human_name 'Vulnerabilities saved per day (per month)'
  set_kind :line_chart

  def chart
    scope = Pundit.policy_scope(current_user, Vulnerability)
    result = scope.group_by_day(
      'vulnerabilities.created_at',
      range: 1.month.ago.midnight..Time.now
    ).count
    [{name: 'Count', data: result}]
  end
end
