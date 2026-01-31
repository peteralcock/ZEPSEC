# frozen_string_literal: true

class FiltredTableVulnersReport < BaseReport
  include DateTimeHelper

  set_lang :en
  set_report_name :filtred_table_vulners
  set_human_name 'Field search results'
  set_report_model 'Vulnerability'
  set_required_params %i[q]
  set_formats %i[csv]

  def csv(blank_document)
    r = blank_document

    header = [
      '№',
      'CVE ID',
      'Other IDs',
      'Processed',
      'Processed by',
      'Processing date',
      'Category',
      'CWE ID',
      'Last data entry source',
      'Source update date',
      'Source publication date',
      'Date updated in database',
      'Date saved to database',
      'Vendors',
      'Products',
      'Versions',
      'Criticality',
      'CVSS base score',
      'CVSS vector',
      'Attack vector',
      'Exploit availability',
      'Relevance',
      'Applicability',
      'Source description',
      'Description',
      'NVD link',
      'Source references',
      'References',
      'Recommendations',
      'Bulletins'
    ]
    custom_fields = CustomField.where(field_model: 'Vulnerability')
    custom_fields_names = custom_fields.each_with_object([]) { |v, o| o << v.name }
    r. << (header + custom_fields_names)

    @records.each_with_index do |record, index|
      row = []
      record = VulnerabilityDecorator.new(record)

      row << index + 1
      row << record.codename
      row << record.show_custom_codenames
      row << record.show_processed
      row << record.processor&.name
      row << show_date_time(record.processed_at)
      row << record.vulnerability_kind&.name
      row << record.show_cwe
      row << record.show_feed
      row << show_date_time(record.modified)
      row << show_date_time(record.published)
      row << show_date_time(record.updated_at)
      row << show_date_time(record.created_at)
      row << record.show_vendors_text
      row << record.show_products_text
      row << record.show_versions_by_products_text
      row << record.show_criticality
      row << record.show_cvss
      row << record.show_cvss_vector
      row << record.show_cvss_av
      row << record.show_custom_exploit_maturity
      row << record.show_custom_actuality
      row << record.show_custom_relevance
      row << record.show_description_string
      row << record.custom_description
      row << record.show_link_nvd_codename
      row << record.show_references_string(separator: "\n")
      row << record.custom_references
      row << record.custom_recomendation
      row << record.show_bulletins_string
      custom_fields.each do |c|
        row << record.custom_field(c.name)
      end

      r << row
    end
  end

  private

  def get_records(options, organization)
    scope = Vulnerability.includes(:vulnerability_bulletins)
    if options[:q].present?
      q = scope.ransack(options[:q])
      q.sorts = options[:q].fetch('s', default_sort)
      q.result.limit(2000)
    else
      scope.all.limit(2000).sort(default_sort)
    end
  end

  def default_sort
    'modified desc'
  end
#
#  def records(scope)
#    scope = policy_scope(scope)
#    @q = scope.ransack(params[:q])
#    @q.sorts = default_sort if @q.sorts.empty?
#    @q.result
#      .includes(records_includes)
#      .page(params[:page])
#  end
end
