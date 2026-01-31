class FailedLoginsByDays < BaseChart
  set_chart_name :failed_logins_by_days
  set_human_name 'Failed login attempts by days (per month)'
  set_kind :line_chart

  def chart
    result = UserAction.where(event: 101)
      .group_by_day(
        :created_at,
        range: 1.month.ago.midnight..Time.now
      ).count
    [{name: 'Count', data: result}]
  end
end
