# frozen_string_literal: true

class OrganizationsReport < BaseReport
  set_lang :en
  set_report_name :organizations
  set_human_name 'Organizations'
  set_report_model 'Organization'
  set_required_params %i[]
  set_formats %i[docx]

  def docx(blank_document)
    r = blank_document
    organizations = OrganizationPolicy::Scope.new(current_user, Organization).resolve
    r.p  "Organizations report", style: 'Header'
    organizations.each_with_index do |organization, index|
      r.p
      r.p "#{index + 1}. #{organization.name}", style: 'Header'
      r.p "Description: #{organization.description}", style: 'Text'
    end
  end

  private

  def get_records(_options, _organization);end
end
