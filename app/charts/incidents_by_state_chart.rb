class IncidentsByStateChart < BaseChart
  set_chart_name :incidents_by_state
  set_human_name 'Incidents by state (closed shown for the last month)'
  set_kind :bar_chart

  def chart
    result = Incident.group(:state).count
    [{name: 'Count', data: result}]
    sql = <<~SQL
      SELECT
        COUNT(*) AS count_all,
        CASE incidents.state
        WHEN 0
          THEN 'Registered'
        WHEN 1
          THEN 'In progress'
        WHEN 2
          THEN 'Closed'
        END
        AS incidents_state
      FROM
        incidents
      WHERE
        incidents.closed_at > NOW() - INTERVAL '30 days'
        OR
        incidents.closed_at IS NULL
      GROUP BY 2
    SQL
    result = Incident.find_by_sql(sql).each_with_object({}) do
      |i, memo| memo[i.incidents_state] = i.count_all
    end
    [{name: 'Count', data: result}]
  end
end
