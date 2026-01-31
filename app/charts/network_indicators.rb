class NetworkIndicators < BaseChart
  set_chart_name :network_indicators
  set_human_name 'Top indicators - network addresses'
  set_kind :pie_chart

  def chart
    scope = Pundit.policy_scope(current_user, Indicator)
    scope.where(content_format: 'network').top(:content, 20)
  end
end
