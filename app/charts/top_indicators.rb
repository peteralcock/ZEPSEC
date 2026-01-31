class TopIndicators < BaseChart
  set_chart_name :top_indicators
  set_human_name 'Indicators by content (top 20)'
  set_kind :column_chart

  def chart
    Indicator.top(:content, 20)
  end
end
