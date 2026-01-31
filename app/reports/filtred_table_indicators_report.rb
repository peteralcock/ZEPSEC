# frozen_string_literal: true

class FiltredTableIndicatorsReport < BaseReport
  include DateTimeHelper

  set_lang :en
  set_report_name :filtred_table_indicators
  set_human_name 'Field search results'
  set_report_model 'Indicator'
  set_required_params %i[q]
  set_formats %i[csv]

  def csv(blank_document)
    r = blank_document

    header = [
      '№',
      'Indicator format',
      'Indicator contexts',
      'Indicator',
      'Investigation type',
      'Investigation code',
      'Investigation source',
      'Investigation source code',
      'Investigation description',
      'Trust level',
      'Purpose',
      'Parent container',
      'Description',
      'Created at',
      'Created by',
      'Updated at',
      'Updated by',
      'Owner organization',
    ]
    custom_fields = CustomField.where(field_model: 'Indicator')
    custom_fields_names = custom_fields.each_with_object([]) { |v, o| o << v.name }
    r. << (header + custom_fields_names)

    @records.each_with_index do |record, index|
      row = []
      record = IndicatorDecorator.new(record)

      row << index + 1
      row << record.show_content_format
      row << record.show_indicator_contexts
      row << record.content
      row << record.investigation.investigation_kind.name
      row << record.investigation.custom_codename
      row << record.investigation.feed.name
      row << record.investigation.feed_codename
      row << record.investigation.description
      row << record.show_trust_level
      row << record.show_purpose
      row << record.parent&.content
      row << record.description
      row << show_date_time(record.created_at)
      row << record.creator.name
      row << show_date_time(record.updated_at)
      row << record.updater&.name
      row << record.organization.name

      custom_fields.each do |c|
        row << record.custom_field(c.name)
      end

      r << row
    end
  end

  private

  def get_records(options, organization)
    scope = Pundit.policy_scope(current_user, Indicator)
      .includes(
        :organization,
        :creator,
        :updater,
        :investigation_kind,
        :indicator_contexts
      )
    if options[:q].present?
      q = scope.ransack(options[:q])
      q.sorts = options[:q].fetch('s', default_sort) #  default_sort if q.sorts.empty?
      q.result.limit(2000)
    else
      scope.all.limit(2000).sort(default_sort)
    end
  end

  def default_sort
    'created_at desc'
  end
end
