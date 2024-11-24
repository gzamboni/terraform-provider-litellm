resource "litellm_team" "example_team" {
  team_id = "my-team-id"
  team_alias = "my-team-id"
  tpm_limit = 2000
  rpm_limit = 20
  max_budget = 100.0
  metadata = {
    "createdBy" = "terraform-provider-litellm"
  }
  models = ["azure/gpt-4o"]
  blocked = false
}