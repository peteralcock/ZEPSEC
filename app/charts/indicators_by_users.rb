class IndicatorsByUsers < BaseChart
  set_chart_name :indicators_by_users
  set_human_name 'Indicators by contributing users'
  set_kind :column_chart

  def chart
    scope = Pundit.policy_scope(current_user, Indicator)
    result = scope.joins(:creator)
      .top('users.name', 10)
    [{name: 'Count', data: result}]
  end
end
