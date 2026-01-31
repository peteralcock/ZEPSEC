class IndicatorsByDays < BaseChart
  set_chart_name :indicators_by_days
  set_human_name 'Indicators by days (per month)'
  set_kind :line_chart

  def chart
    scope = Pundit.policy_scope(current_user, Indicator)
    result = scope.group_by_day('indicators.created_at', range: 1.month.ago.midnight..Time.now).count
    [{name: 'Count', data: result}]
  end
end
