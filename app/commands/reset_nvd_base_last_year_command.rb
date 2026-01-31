# frozen_string_literal: true

class ResetNvdBaseLastYearCommand < BaseCommand
  set_command_name :reset_nvd_base_last_year
  set_human_name 'Recreate NVD data for the last year'
  set_command_model 'Vulnerability'
  set_required_params %i[]

  def run
    return unless @current_user.admin?
    year = Time.current.year
    Vulnerability.where(feed: Vulnerability.feeds[:nvd])
                 .where('codename LIKE :codename', codename: "CVE-#{year}%")
                 .delete_all
    ResetNvdBaseJob.perform_later('sync_nvd_base', year)
  end
end
